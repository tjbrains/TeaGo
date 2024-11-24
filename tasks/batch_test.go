package tasks_test

import (
	"github.com/tjbrains/TeaGo/tasks"
	"testing"
)

func TestBatch_Run(t *testing.T) {
	var b = tasks.NewBatch()
	b.Add(func() {
		t.Log("1")
	})
	b.Add(func() {
		t.Log("2")
	})
	b.Add(func() {
		t.Log("3")
	})
	b.Add(func() {
		t.Log("4")
	})
	b.Run()
	t.Log("done")
}
