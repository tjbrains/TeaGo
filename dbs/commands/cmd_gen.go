package commands

import (
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tjbrains/TeaGo/Tea"
	"github.com/tjbrains/TeaGo/cmd"
	"github.com/tjbrains/TeaGo/dbs"
	"github.com/tjbrains/TeaGo/files"
	"github.com/tjbrains/TeaGo/lists"
	"github.com/tjbrains/TeaGo/utils/string"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type GenModelCommand struct {
	*cmd.Command
}

func (this *GenModelCommand) Name() string {
	return "generate model and dao files"
}

func (this *GenModelCommand) Codes() []string {
	return []string{":db.gen"}
}

func (this *GenModelCommand) Usage() string {
	return ":db.gen [MODEL_NAME] [-db=[DB ID] -dir=[TARGET DIR]]"
}

func (this *GenModelCommand) Run() {
	model, found := this.Arg(1)
	if !found {
		this.Error(errors.New("please specify model name"))
		return
	}
	dbId, found := this.Param("db")
	var db *dbs.DB
	var err error
	if found {
		db, err = dbs.Instance(dbId)
		if err != nil {
			this.Error(err)
			return
		}
	} else {
		db, err = dbs.Default()
		if err != nil {
			this.Error(err)
			return
		}
	}

	// 模型目录
	subPackage := "models"
	dir, _ := this.Param("dir")
	config, _ := db.Config()
	if len(config.Models.Package) > 0 {
		dir = strings.TrimSuffix(config.Models.Package+Tea.DS+dir, Tea.DS)
	}

	packagePieces := strings.Split(model, ".")
	if len(packagePieces) > 0 {
		model = packagePieces[len(packagePieces)-1]

		if len(dir) > 0 {
			dir += Tea.DS + strings.Join(packagePieces[:len(packagePieces)-1], Tea.DS)

			dirFile := files.NewFile(dir)
			if !dirFile.Exists() {
				err := dirFile.MkdirAll()
				if err != nil {
					this.Println(err.Error())
					return
				}
			}
		} else {
			subPackage = packagePieces[len(packagePieces)-2]
		}
	}

	if len(dir) > 0 {
		subPackage = filepath.Base(dir)
	}

	// 取得对应表
	subTableName, err := this.modelToTable(model)
	if err != nil {
		this.Error(err)
		return
	}
	tableName := db.TablePrefix() + subTableName

	tableNames, err := db.TableNames()
	if err != nil {
		this.Error(err)
		return
	}
	lowerTableName := strings.Replace(strings.ToLower(tableName), "_", "", -1)
	for _, dbTableName := range tableNames {
		if strings.ToLower(strings.Replace(dbTableName, "_", "", -1)) == lowerTableName {
			tableName = dbTableName
			break
		}
	}

	table, err := db.FindTable(tableName)
	if err != nil {
		this.Error(err)
		return
	}
	if table == nil {
		this.Println("can not find table named '" + tableName + "'")
		return
	}

	// Model
	var modelString = `package ` + subPackage
	modelString += `

import "github.com/tjbrains/TeaGo/dbs"`

	// fields
	modelString += "\n"
	modelString += "const (\n"
	for _, field := range table.Fields {
		var attr = this.convertFieldNameStyle(field.Name)
		modelString += "\t" + model + "Field_" + attr + " dbs.FieldName = \"" + field.Name + "\" // " + field.Comment + "\n"
	}
	modelString += ")\n"

	modelString += `

// ` + model + " " + strings.Replace(table.Comment, "\n", " ", -1) + `
type ` + model + ` struct {`
	modelString += "\n"
	var primaryKey = ""
	var primaryKeyType = ""
	fieldNames := []string{}
	for _, field := range table.Fields {
		fieldNames = append(fieldNames, field.Name)

		if field.IsPrimaryKey {
			primaryKeyType = field.ValueTypeName()
			primaryKey = field.Name
		}
	}

	var specialBoolFields = []string{}
	var globalConfig = dbs.GlobalConfig()
	if globalConfig.Fields != nil {
		specialBoolFields = globalConfig.Fields["bool"]
	}

	for _, field := range table.Fields {
		var attr = this.convertFieldNameStyle(field.Name)
		var dataType = field.ValueTypeName()

		// bool类型
		if lists.ContainsString(specialBoolFields, field.Name) || this.isBoolField(field.Name) {
			dataType = "bool"
		}

		modelString += "\t" + attr + " " + dataType + " `field:\"" + field.Name + "\"` // " + field.Comment + "\n"
	}

	modelString += "}"
	modelString += "\n\n"

	// Operator
	modelString += `type ` + model + `Operator struct {`
	modelString += "\n"

	for _, field := range table.Fields {
		var attr = this.convertFieldNameStyle(field.Name)
		modelString += "\t" + attr + " any" + " // " + field.Comment + "\n"
	}

	modelString += "}"
	modelString += "\n"
	modelString += `
func New` + model + `Operator() *` + model + `Operator {
	return &` + model + `Operator{}
}
`

	formatted, err := format.Source([]byte(modelString))
	if err == nil {
		modelString = string(formatted)
	}

	if len(dir) == 0 {
		fmt.Println("Model:")
		fmt.Println("~~~")
		fmt.Println(modelString)
		fmt.Println("~~~")
	} else {
		// 写入文件
		target := filepath.Clean(os.Getenv("GOPATH") + Tea.DS + dir + Tea.DS + this.convertToUnderlineName(model) + "_model.go")
		file := files.NewFile(target)
		if file.Exists() && !this.HasParam("force") {
			this.Output("<error>write failed: '" + strings.TrimPrefix(target, os.Getenv("GOPATH")) + "' already exists</error>\n")
		} else {
			err := file.WriteString(modelString)
			if err != nil {
				this.Error(err)
			} else {
				this.Output("<ok>write '" + strings.TrimPrefix(target, os.Getenv("GOPATH")) + "' ok</ok>\n")
			}
		}
	}

	// model ext
	extString := `package ` + subPackage + `
`
	formatted, err = format.Source([]byte(extString))
	if err == nil {
		extString = string(formatted)
	}

	if len(dir) == 0 {
		fmt.Println("Model Ext:")
		fmt.Println("~~~")
		fmt.Println(extString)
		fmt.Println("~~~")
	} else {
		// 写入文件
		target := filepath.Clean(os.Getenv("GOPATH") + Tea.DS + dir + Tea.DS + this.convertToUnderlineName(model) + "_model_ext.go")
		file := files.NewFile(target)
		if file.Exists() {
			this.Output("<error>write failed: '" + strings.TrimPrefix(target, os.Getenv("GOPATH")) + "' already exists</error>\n")
		} else {
			err := file.WriteString(extString)
			if err != nil {
				this.Error(err)
			} else {
				this.Output("<ok>write '" + strings.TrimPrefix(target, os.Getenv("GOPATH")) + "' ok</ok>\n")
			}
		}
	}

	// DAO
	daoString := `package ` + subPackage + `

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/tjbrains/TeaGo/dbs"
	"github.com/tjbrains/TeaGo/Tea"
)
`

	if stringutil.Contains(fieldNames, "state") {
		daoString += `const (
	${model}StateEnabled = 1 // 已启用
	${model}StateDisabled = 0 // 已禁用
)
`
	}

	daoString += `type ` + model + `DAO dbs.DAO

func New` + model + `DAO() *` + model + `DAO {
	return dbs.NewDAO(&` + model + `DAO{
		DAOObject: dbs.DAOObject{
			DB:     Tea.Env,
			Table:  "` + tableName + `",
			Model:  new(` + model + `),
			PkName: "` + primaryKey + `",
		},
	}).(*` + model + `DAO)
}

var Shared` + model + `DAO *` + model + `DAO

func init() {
	dbs.OnReady(func () {
		Shared` + model + `DAO = New` + model + `DAO()
	})
}

`
	// state
	if stringutil.Contains(fieldNames, "state") {
		daoString += `
// Enable${model} 启用条目
func (this *${daoName}) Enable${model}(tx *dbs.Tx, ${pkName} ${pkNameType}) error {
	_, err := this.Query(tx).
		Pk(${pkName}).
		Set("state", ${model}StateEnabled).
		Update()
	return err
}

// Disable${model} 禁用条目
func (this *${daoName}) Disable${model}(tx *dbs.Tx, ${pkName} ${pkNameType}) error {
	_, err := this.Query(tx).
		Pk(${pkName}).
		Set("state", ${model}StateDisabled).
		Update()
	return err
}

// FindEnabled${model} 查找启用中的条目
func (this *${daoName}) FindEnabled${model}(tx *dbs.Tx, ${pkName} ${pkNameType}) (*${model}, error) {
	result, err := this.Query(tx).
		Pk(${pkName}).
		State(${model}StateEnabled).
		Find()
	if result == nil {
		return nil, err
	}
	return result.(*${model}), err
}
`
	}

	if stringutil.Contains(fieldNames, "name") && table.FindFieldWithName("name").ValueTypeName() == "string" {
		daoString += `// Find${model}Name 根据主键查找名称
func (this *${daoName}) Find${model}Name(tx *dbs.Tx, ${pkName} ${pkNameType}) (string, error) {
	return this.Query(tx).
		Pk(${pkName}).
		Result("name").
		FindStringCol("")
}
`
	}

	daoString = strings.Replace(daoString, "${daoName}", model+"DAO", -1)
	daoString = strings.Replace(daoString, "${pkName}", primaryKey, -1)
	daoString = strings.Replace(daoString, "${pkNameType}", primaryKeyType, -1)
	daoString = strings.Replace(daoString, "${model}", model, -1)

	formatted, err = format.Source([]byte(daoString))
	if err == nil {
		daoString = string(formatted)
	}

	if len(dir) == 0 {
		fmt.Print("\n\n")
		fmt.Println("DAO:")
		fmt.Println("~~~")
		fmt.Println(daoString)
		fmt.Println("~~~")
	} else {
		// 写入文件
		target := filepath.Clean(os.Getenv("GOPATH") + Tea.DS + dir + Tea.DS + this.convertToUnderlineName(model) + "_dao.go")
		file := files.NewFile(target)
		if file.Exists() {
			this.Output("<error>write failed: '" + strings.TrimPrefix(target, os.Getenv("GOPATH")) + "' already exists</error>\n")
		} else {
			err := file.WriteString(daoString)
			if err != nil {
				this.Error(err)
			} else {
				this.Output("<ok>write '" + strings.TrimPrefix(target, os.Getenv("GOPATH")) + "' ok</ok>\n")
			}
		}
	}

	// test
	testString := `package ` + subPackage + `_test
import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/tjbrains/TeaGo/bootstrap"
)
`
	formatted, err = format.Source([]byte(testString))
	if err == nil {
		testString = string(formatted)
	}

	if len(dir) == 0 {
		fmt.Print("\n\n")
		fmt.Println("DAO Test:")
		fmt.Println("~~~")
		fmt.Println(testString)
		fmt.Println("~~~")
	} else {
		// 写入文件
		target := filepath.Clean(os.Getenv("GOPATH") + Tea.DS + dir + Tea.DS + this.convertToUnderlineName(model) + "_dao_test.go")
		file := files.NewFile(target)
		if file.Exists() {
			this.Output("<error>write failed: '" + strings.TrimPrefix(target, os.Getenv("GOPATH")) + "' already exists</error>\n")
		} else {
			err := file.WriteString(testString)
			if err != nil {
				this.Error(err)
			} else {
				this.Output("<ok>write '" + strings.TrimPrefix(target, os.Getenv("GOPATH")) + "' ok</ok>\n")
			}
		}
	}
}

func (this *GenModelCommand) modelToTable(modelName string) (string, error) {
	var tableName = modelName + "s"

	// ies
	if strings.HasSuffix(tableName, "ys") && !regexp.MustCompile("[aeiou]ys$").MatchString(tableName) {
		tableName = tableName[:len(tableName)-2] + "ies"
	}

	// oes
	reg, err := stringutil.RegexpCompile("(?i)(hero|potato|tomato|echo|tornado|torpedo|domino|veto|mosquito|negro|mango|buffalo|volcano|match|dish|brush|branch|dress|glass|bus|class|boss|process|box|fox|watch|index)s")
	if err != nil {
		return tableName, err
	}
	tableName = reg.ReplaceAllString(tableName, "${1}es")

	// ves
	for find, replace := range map[string]string{
		"leafs$":          "leaves",
		"halfs$":          "halves",
		"wolfs$":          "wolves",
		"shiefs$":         "shieves",
		"shelfs$":         "shelves",
		"knifes$":         "knives",
		"wifes$":          "wives",
		"(goods|money)s$": "$1",
	} {
		reg, err = stringutil.RegexpCompile("(?i)" + find)
		if err != nil {
			return tableName, err
		}
		tableName = reg.ReplaceAllString(tableName, replace)
	}

	return tableName, nil
}

func (this *GenModelCommand) convertFieldNameStyle(fieldName string) string {
	pieces := strings.Split(fieldName, "_")
	newPieces := []string{}
	for _, piece := range pieces {
		newPieces = append(newPieces, strings.ToUpper(string(piece[0]))+piece[1:])
	}
	return strings.Join(newPieces, "")
}

func (this *GenModelCommand) convertToUnderlineName(modelName string) string {
	// 如果名字前面有多个大写字母则认为是同一个单词
	reg := regexp.MustCompile(`[A-Z]{2,}`)
	modelName = reg.ReplaceAllStringFunc(modelName, func(s string) string {
		return "_" + strings.ToLower(s[:len(s)-1]) + s[len(s)-1:]
	})

	// 将单个大写字母转换为"下划线_小写"
	reg = regexp.MustCompile(`[A-Z]`)
	return strings.TrimPrefix(reg.ReplaceAllStringFunc(modelName, func(s string) string {
		return "_" + strings.ToLower(s)
	}), "_")
}

func (this *GenModelCommand) isBoolField(fieldName string) bool {
	for _, prefix := range []string{"is", "can", "has", "should"} {
		if strings.HasPrefix(fieldName, prefix) && len(fieldName) > len(prefix) && (fieldName[len(prefix)] >= 'A' && fieldName[len(prefix)] <= 'Z') {
			return true
		}
	}

	return false
}
