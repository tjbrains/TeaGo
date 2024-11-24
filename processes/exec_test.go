package processes_test

import (
	"github.com/tjbrains/TeaGo/processes"
	"testing"
)

func TestExecAndReturn(t *testing.T) {
	t.Log(processes.Exec("/usr/local/bin/php", "-v"))
}
