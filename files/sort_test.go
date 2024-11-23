package files_test

import (
	"github.com/tjbrains/TeaGo/files"
	"os"
	"testing"
)

func TestSortSize(t *testing.T) {
	var fileObject = files.NewFile(os.Getenv("GOPATH") + "/src/github.com/tjbrains/TeaGo").List()

	files.Sort(fileObject, files.SortTypeSize)

	for _, file := range fileObject {
		size, _ := file.Size()
		if file.IsDir() {
			t.Log("d:"+file.Name(), size)
		} else {
			t.Log(file.Name(), size)
		}
	}
}

func TestSortKind(t *testing.T) {
	var fileObject = files.NewFile(os.Getenv("GOPATH") + "/src/github.com/tjbrains/TeaGo").List()

	files.Sort(fileObject, files.SortTypeKind)

	for _, file := range fileObject {
		if file.IsDir() {
			t.Log("d:" + file.Name())
		} else {
			t.Log(file.Name())
		}
	}
}

func TestSortKindReverse(t *testing.T) {
	var fileObject = files.NewFile(os.Getenv("GOPATH") + "/src/github.com/tjbrains/TeaGo").List()

	files.Sort(fileObject, files.SortTypeKindReverse)

	for _, file := range fileObject {
		if file.IsDir() {
			t.Log("d:" + file.Name())
		} else {
			t.Log(file.Name())
		}
	}
}
