package rands_test

import (
	"github.com/tjbrains/TeaGo/rands"
	"runtime"
	"testing"
)

func TestRand_Distribute_1(t *testing.T) {
	var m = map[int]int{} // number => count
	for i := 0; i < 1000000; i++ {
		var v = rands.Int(0, 9)
		_, ok := m[v]
		if ok {
			m[v]++
		} else {
			m[v] = 1
		}
	}
	t.Log(m)
}

func TestRand_Distribute_2(t *testing.T) {
	var m = map[int]int{} // number => count
	for i := 0; i < 1000000; i++ {
		v := rands.Int(15, 5)
		_, ok := m[v]
		if ok {
			m[v]++
		} else {
			m[v] = 1
		}
	}
	t.Log(m)
}

func BenchmarkRandBetween(b *testing.B) {
	runtime.GOMAXPROCS(1)

	for i := 0; i < b.N; i++ {
		_ = rands.Int(0, 100)
	}
}

func BenchmarkRandBetween_Concurrent(b *testing.B) {
	runtime.GOMAXPROCS(1)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = rands.Int(0, 100)
		}
	})
}
