package commands

import (
	"testing"
	"github.com/tjbrains/TeaGo/cmd"
)

func TestExecCommand_Run(t *testing.T) {
	cmd.Try([]string{ ":db.exec", "UPDATE pp_adHelps SET `order`=1", "-db=dev" })
}
