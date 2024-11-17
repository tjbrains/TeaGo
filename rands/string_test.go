package rands_test

import (
	"github.com/tjbrains/TeaGo/rands"
	"runtime"
	"testing"
)

func TestRand_String(t *testing.T) {
	t.Log(rands.String(32))
	t.Log(rands.String(32))
	t.Log(rands.String(32))
	t.Log(rands.String(0))
	t.Log(rands.String(64))
}

func TestRand_HexString(t *testing.T) {
	t.Log(rands.HexString(32))
	t.Log(rands.HexString(32))
	t.Log(rands.HexString(32))
	t.Log(rands.HexString(0))
	t.Log(rands.HexString(64))
}

func TestRand_UniqueString(t *testing.T) {
	var m = map[string]bool{}
	for i := 0; i < 10_000_000; i++ {
		s := rands.String(32)
		_, ok := m[s]
		if ok {
			t.Fatal("duplicated:", s)
		}
		m[s] = true
	}
	t.Log("ok")
}

func BenchmarkRand_String(b *testing.B) {
	runtime.GOMAXPROCS(1)

	for i := 0; i < b.N; i++ {
		_ = rands.String(32)
	}
}

func BenchmarkRand_String_Concurrent(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rands.String(32)
		}
	})
}

func BenchmarkRand_HexString(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = rands.HexString(32)
	}
}

func BenchmarkRand_HexString_Concurrent(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rands.HexString(32)
		}
	})
}
