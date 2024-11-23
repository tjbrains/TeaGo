package files_test

import (
	"github.com/tjbrains/TeaGo/Tea"
	"github.com/tjbrains/TeaGo/files"
	"testing"
)

func TestWriter_Write(t *testing.T) {
	var tmpFile = files.NewFile(Tea.TmpFile("test.txt"))
	writer, err := tmpFile.Writer()
	if err != nil {
		t.Fatal(err)
	}

	//writer.Write([]byte("Hello,a"))
	//writer.Truncate()

	//writer.Seek(10)
	//writer.WriteString("ba")

	writer.Close()
}
