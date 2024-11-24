package processes_test

import (
	"github.com/tjbrains/TeaGo/processes"
	"testing"
)

func TestNewProcess(t *testing.T) {
	var process = processes.NewProcess("/usr/local/bin/php", "-v")
	err := process.Start()
	if err != nil {
		t.Fatal("[ERROR]", err)
	}

	t.Log(process.Pid())

	//process.Wait()
}
