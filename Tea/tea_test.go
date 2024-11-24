package Tea_test

import (
	"github.com/tjbrains/TeaGo/Tea"
	"github.com/tjbrains/TeaGo/assert"
	"os"
	"testing"
)

func TestFindLatestDir(t *testing.T) {
	t.Log(Tea.Root)
}

func TestTmpDir(t *testing.T) {
	t.Log(Tea.TmpDir())
	t.Log(Tea.TmpFile("test.json"))
}

func TestIsTesting(t *testing.T) {
	a := assert.NewAssertion(t).Quiet()
	a.IsTrue(Tea.IsTesting())
	t.Log(os.Args)
}
