package files_test

import (
	"github.com/tjbrains/TeaGo/Tea"
	"github.com/tjbrains/TeaGo/files"
	"testing"
)

func Test_Appender(t *testing.T) {
	var tmpFile = files.NewFile(Tea.TmpFile("test.txt"))
	appender, err := tmpFile.Appender()
	if err != nil {
		t.Fatal(err)
	}

	//appender.Lock()
	appender.Append([]byte("Hello,a"))
	//appender.Truncate()

	appender.AppendString("[ABC]")

	//appender.Unlock()
	t.Log(appender.Close())
}
