package commands

import (
	"github.com/tjbrains/TeaGo/cmd"
	"github.com/tjbrains/TeaGo/dbs"
	"gopkg.in/yaml.v3"
)

type InfoCommand struct {
	*cmd.Command
}

func (this *InfoCommand) Name() string {
	return "print database info"
}

func (this *InfoCommand) Codes() []string {
	return []string{":db.info"}
}

func (this *InfoCommand) Usage() string {
	return ":db.info"
}

func (this *InfoCommand) Run() {
	db, err := dbs.Default()
	if err != nil {
		this.Error(err)
		return
	}

	config, _ := db.Config()
	yamlBytes, err := yaml.Marshal(config)
	if err != nil {
		this.Error(err)
		return
	}

	this.Output("<code>" + string(yamlBytes) + "</code>")
}
