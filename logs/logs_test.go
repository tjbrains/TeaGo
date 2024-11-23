package logs_test

import (
	"errors"
	"github.com/tjbrains/TeaGo/logs"
	"testing"
)

func TestDump(t *testing.T) {
	var m = map[string]interface{}{
		"name": "Liu",
		"age":  20,
		"book": map[string]interface{}{
			"name":  "Golang",
			"price": 20.00,
		},
	}
	logs.Dump(m)
}

func TestError(t *testing.T) {
	logs.Error(errors.New("This is error!!!"))
}
