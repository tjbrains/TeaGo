package commands

import (
	"github.com/tjbrains/TeaGo/cmd"
	"testing"
)

func TestCompareDBCommand_Run(t *testing.T) {
	cmd.Try([]string{":db.compare", "dev", "remote"})
}
