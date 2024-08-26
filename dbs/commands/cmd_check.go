package commands

import (
	"github.com/tjbrains/TeaGo/Tea"
	"github.com/tjbrains/TeaGo/cmd"
	"github.com/tjbrains/TeaGo/dbs"
	"github.com/tjbrains/TeaGo/files"
	"github.com/tjbrains/TeaGo/lists"
	"os"
	"regexp"
	"strings"
)

type CheckModelCommand struct {
	*cmd.Command
}

func (this *CheckModelCommand) Name() string {
	return "check model's modification"
}

func (this *CheckModelCommand) Codes() []string {
	return []string{":db.check"}
}

func (this *CheckModelCommand) Usage() string {
	return ":db.check"
}

func (this *CheckModelCommand) Run() {
	// 所有的模型
	db, err := dbs.Default()
	if err != nil {
		this.Error(err)
		return
	}

	config, err := db.Config()
	if err != nil {
		this.Error(err)
		return
	}

	pkg := config.Models.Package
	if len(pkg) == 0 {
		this.Println("'models.package' should be configured for db '" + db.Id() + "'")
		return
	}

	dir := files.NewFile(os.Getenv("GOPATH") + Tea.DS + pkg)
	if !dir.Exists() {
		this.Println("'" + pkg + "' does not exist")
		return
	}

	this.Output("<code>checking ...</code>\n~~~\n")

	var tables = []*dbs.Table{}                     // Model name => *Table
	var models = map[string]map[string]*dbs.Field{} // Model name => { fields:... }
	var modelFiles = map[string]string{}            // Model name => File name
	var countIssues = 0

	var specialBoolFields = []string{}
	var globalConfig = dbs.GlobalConfig()
	if globalConfig.Fields != nil {
		specialBoolFields = globalConfig.Fields["bool"]
	}

	dir.Range(func(file *files.File) {
		if !file.IsFile() {
			return
		}

		if !strings.HasSuffix(file.Name(), ".go") {
			return
		}

		content, err := file.ReadAllString()
		if err != nil {
			this.Error(err)
			return
		}

		content = strings.Replace(content, "\n", " ", -1)
		content = strings.Replace(content, "\r", " ", -1)

		// DAO
		reg := regexp.MustCompile("dbs.DAOObject{\\s*DB:.+,\\s*Table:\\s*\"(\\w+)\",\\s*Model:\\s*new\\((\\w+)\\),\\s*PkName:\\s*\"(\\w+)\"")
		if reg.MatchString(content) {
			match := reg.FindStringSubmatch(content)[1:]
			tableName := match[0]
			modelName := match[1]

			// 表信息
			table, err := db.FindTable(tableName)
			if err != nil || table == nil {
				path, _ := file.AbsPath()
				this.Output("<code>-[" + modelName + "] remove model</code>\n")
				this.outputFile(path)

				countIssues++
				return
			}

			table.MappingName = modelName
			tables = append(tables, table)
		}

		reg = regexp.MustCompile("type\\s+(\\w+)\\s+struct {.+}")
		if reg.MatchString(content) {
			match := reg.FindStringSubmatch(content)[1:]
			modelName := match[0]

			modelFiles[modelName], _ = file.AbsPath()

			// 所有字段
			reg = regexp.MustCompile("(\\w+)\\s+(\\w+|\\[]byte|dbs\\.JSON)\\s*`field:\"(\\w+)\"`")
			matches := reg.FindAllStringSubmatch(content, -1)
			fields := map[string]*dbs.Field{}
			for _, match := range matches {
				var mappingName = match[1]
				var dataTypeString = match[2]
				var fieldName = match[3]

				var field = new(dbs.Field)
				field.Name = fieldName
				field.MappingName = mappingName
				field.MappingKindName = dataTypeString

				fields[field.Name] = field
			}

			models[modelName] = fields
		}
	})

	// 检查现有table
	for _, table := range tables {
		modelName := table.MappingName
		oldFields, found := models[modelName]
		if !found {
			this.Output("+[" + modelName + "] gen model\n")
		} else {
			// 新增字段或修改字段
			for _, field := range table.Fields {
				var newTypeName = field.ValueTypeName()

				// bool类型
				if lists.ContainsString(specialBoolFields, field.Name) || this.isBoolField(field.Name) {
					newTypeName = "bool"
				}

				oldField, found := oldFields[field.Name]
				if !found {
					this.Output("<code>+[" + modelName + "] field: " + this.convertFieldNameStyle(field.Name) + " " + newTypeName + " `field:\"" + field.Name + "\"` // " + field.Comment + "</code>\n")
					this.Output("<code>+[" + modelName + "Operator] field: " + this.convertFieldNameStyle(field.Name) + " any // " + field.Comment + "</code>\n")
					this.outputFile(modelFiles[modelName])
					countIssues++
				} else {
					// 对比
					if newTypeName != oldField.MappingKindName {
						this.Output("<code>*[" + modelName + "] field: " + oldField.MappingName + " " + newTypeName + " `field:\"" + field.Name + "\"` // " + field.Comment + "</code>\n")
						this.outputFile(modelFiles[modelName])
						countIssues++
					}
				}
			}

			// 删除字段
			for _, oldField := range oldFields {
				field := table.FindFieldWithName(oldField.Name)
				if field == nil {
					this.Output("<code>-[" + modelName + "] field: " + oldField.MappingName + "</code>\n")
					this.outputFile(modelFiles[modelName])
					countIssues++
				}
			}
		}
	}

	if countIssues == 0 {
		this.Output("<ok>Everything goes ok</ok>\n")
	} else {
		this.Output("~~~\n")
		this.Output("<error>There are", countIssues, "issues to be fixed</error>\n")
	}
}

func (this *CheckModelCommand) convertFieldNameStyle(fieldName string) string {
	pieces := strings.Split(fieldName, "_")
	newPieces := []string{}
	for _, piece := range pieces {
		newPieces = append(newPieces, strings.ToUpper(string(piece[0]))+string(piece[1:]))
	}
	return strings.Join(newPieces, "")
}

func (this *CheckModelCommand) outputFile(file string) {
	goPath := os.Getenv("GOPATH")
	this.Output("   ", strings.TrimPrefix(file, goPath), "\n")
}

func (this *CheckModelCommand) isBoolField(fieldName string) bool {
	for _, prefix := range []string{"is", "can", "has", "should"} {
		if strings.HasPrefix(fieldName, prefix) && len(fieldName) > len(prefix) && (fieldName[len(prefix)] >= 'A' && fieldName[len(prefix)] <= 'Z') {
			return true
		}
	}

	return false
}
