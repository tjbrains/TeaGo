package logs_test

import (
	"errors"
	"testing"

	"github.com/tjbrains/TeaGo/logs"
)

func TestDump(t *testing.T) {
	var m = map[string]any{
		"name": "Liu",
		"age":  20,
		"book": map[string]any{
			"name":  "Golang",
			"price": 20.00,
		},
	}
	logs.Dump(m)
}

func TestError(t *testing.T) {
	logs.Error(errors.New("This is error!!!"))
}
