package actions_test

import (
	"github.com/tjbrains/TeaGo/actions"
	"testing"
)

func TestData(t *testing.T) {
	var data = actions.Data{}
	data["a"] = "b"
	t.Log(data)
}
