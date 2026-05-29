package dbs

import (
	"testing"
)

func TestMakeModel(t *testing.T) {
	type User struct {
		Id        int `field:"id"`
		Gender    int
		Age       int
		Nickname  string
		CreatedAt int `field:"created_at"`
	}

	var model = NewModel(new(User))
	t.Logf("%#v", model)
}

func BenchmarkModel_New(b *testing.B) {
	type User struct {
		Id        int `field:"id"`
		Gender    int
		Age       int
		Nickname  string
		CreatedAt int `field:"created_at"`
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_ = NewModel(new(User))
	}
}
