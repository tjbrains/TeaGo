package files_test

import (
	"github.com/tjbrains/TeaGo/Tea"
	"github.com/tjbrains/TeaGo/files"
	"os"
	"syscall"
	"testing"
)

func TestFile_Stat(t *testing.T) {
	var file = files.NewFile(Tea.TmpFile("test.txt"))
	stat, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%#v", stat)
	t.Log(stat.ModTime)

	info, _ := os.Stat(Tea.TmpFile("test.txt"))
	t.Logf("%#v", info.Sys().(*syscall.Stat_t).Atimespec)
}

func TestFile_IsFile(t *testing.T) {
	var file = files.NewFile("file.go")
	t.Log(file.Exists())
	t.Log(file.IsFile())
	t.Log(file.IsDir())
}

func TestFile_Read(t *testing.T) {
	var file = files.NewFile("file.go")
	t.Log(file.ReadAll())
	//t.Log(file.ReadAllAsString())
	t.Log(file.MD5())
	t.Log(file.Ext())
}

func TestFile_MkdirAll(t *testing.T) {
	var file = files.NewFile(Tea.Root + "/tmp/a/b/c")
	t.Log(file.MkdirAll())
}

func TestFile_Delete(t *testing.T) {
	var file = files.NewFile(Tea.TmpFile("test.txt"))
	t.Log(file.Delete())
	if file.Exists() {
		t.Fatal("[ERROR]", "delete failed")
	}
}

func TestFile_DeleteDir(t *testing.T) {
	var file = files.NewFile(Tea.TmpFile("test"))
	t.Log(file.DeleteAll())
	if file.Exists() {
		t.Fatal("[ERROR]", "delete failed")
	}
}

func TestFile_List(t *testing.T) {
	var result = files.NewFile("../").List()
	for _, file := range result {
		absPath, _ := file.AbsPath()
		t.Log(file.Name(), file.IsFile(), absPath)
	}
}

func TestFile_Create(t *testing.T) {
	var file = files.NewFile(Tea.TmpFile("test.tmp"))
	t.Log(file.Create())
}

func TestFile_Touch(t *testing.T) {
	var file = files.NewFile(Tea.TmpFile("test.txt"))
	err := file.Touch()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(file.LastModified())
}

func TestFile_Append(t *testing.T) {
	var file = files.NewFile(Tea.TmpFile("test.txt"))
	err := file.Append([]byte("\n"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(file.AppendString("aaaa"))
}

func TestFile_Write(t *testing.T) {
	var file = files.NewFile(Tea.TmpFile("test.txt"))
	t.Log(file.WriteString("aaaa"))
}

func TestFile_Range(t *testing.T) {
	var dir = files.NewFile("../../")
	dir.Range(func(file *files.File) {
		t.Log(file.Path())
	})
}
