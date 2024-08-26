package commands

import (
	"github.com/tjbrains/TeaGo/cmd"
)

func init() {
	cmd.Register(&GenModelCommand{})
	cmd.Register(&CheckModelCommand{})
	cmd.Register(&CompareDBCommand{})
	cmd.Register(&FixCommand{})
	cmd.Register(&SecretCommand{})
	cmd.Register(&InfoCommand{})
	cmd.Register(&ExecCommand{})
	cmd.Register(&ListModelsCommand{})
}
