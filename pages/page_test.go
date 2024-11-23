package pages_test

import (
	"github.com/tjbrains/TeaGo/pages"
	"testing"
)

func TestPageInit(t *testing.T) {
	var page = pages.NewPage(100, 30, 2)
	t.Logf("size:%d, length:%d, offset:%d, index:%d", page.Size, page.Length, page.Offset, page.Index)
}
