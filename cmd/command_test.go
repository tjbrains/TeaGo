package cmd_test

import (
	"github.com/tjbrains/TeaGo/cmd"
	"testing"
)

type testCommand struct {
	*cmd.Command
}

func (command *testCommand) Codes() []string {
	return []string{"test"}
}

func (command *testCommand) Run() {
	command.Println("Run Command")
}

func TestRegister(t *testing.T) {
	var command = &testCommand{}
	cmd.Register(command)
	t.Log(cmd.Run("test"))
}
