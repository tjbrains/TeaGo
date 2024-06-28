package files

import "github.com/tjbrains/TeaGo/Tea"

func NewTmpFile(file string) *File {
	return NewFile(Tea.TmpFile(file))
}
