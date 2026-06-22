// Copyright 2026 FlexCDN root@flexcdn.cn. All rights reserved. Official site: https://flexcdn.cn .

package containers_test

import (
	"math/rand"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/tjbrains/TeaGo/assert"
	"github.com/tjbrains/TeaGo/containers"
	"github.com/tjbrains/TeaGo/types"
)

func TestNewSet(t *testing.T) {
	var a = assert.NewAssertion(t)

	var set = containers.NewSet[uint64, uint64](10)
	set.Push(1, 11)
	set.Push(2, 22)
	set.Push(3, 33)
	set.Push(4, 44)
	set.Push(5, 55)
	set.Push(6, 66)
	set.Push(7, 7)
	set.Push(8, 66)
	set.Inspect(t)

	{
		value, ok := set.Value(3)
		a.IsTrue(ok)
		a.IsTrue(value == 33)
	}

	{
		value, ok := set.Pop()
		a.IsTrue(ok)
		a.IsTrue(value == 7)
	}

	set.Inspect(t)

	t.Log(set.Len(), "keys")

	{
		set.Delete(6)
		set.Delete(8)
	}

	set.Inspect(t)

	{
		_, ok := set.Value(6)
		a.IsFalse(ok)
	}

	set.Push(9, 99)
	set.Push(10, 100)
	set.Push(11, 110)
	set.Push(12, 120)
	set.Push(13, 130)
	set.Push(14, 140)
	a.IsTrue(set.Len() == 10)
	a.IsTrue(set.Contains(13))
	a.IsFalse(set.Contains(200))

	set.Inspect(t)
}

func TestNewSet_NoLimit(t *testing.T) {
	var a = assert.NewAssertion(t)

	var set = containers.NewSet[int, uint64](containers.NoLimit)
	for i := range 1_000_000 {
		set.Push(i, 1)
	}
	a.IsTrue(set.Len() == 1_000_000)
	t.Log("no limit:", containers.NoLimit)
}

func TestSet_Delete(t *testing.T) {
	var a = assert.NewAssertion(t)

	var set = containers.NewSet[int, int](10)
	set.Push(1, 11)
	set.Push(2, 22)
	set.Push(3, 33)
	set.Push(4, 44)
	set.Push(5, 55)

	set.Delete(1, 2, 3)
	a.IsFalse(set.Contains(1))
	a.IsFalse(set.Contains(2))
	a.IsFalse(set.Contains(3))

	set.Inspect(t)

	set.Delete(4, 5)
	a.IsTrue(set.Len() == 0)
	set.Inspect(t)
}

func TestSet_Scan(t *testing.T) {
	var set = containers.NewSet[uint64, int](10)
	for i := range 10 {
		set.Push(uint64(i), rand.Int()/1_000_000)
	}
	var count int
	set.Scan(func(k uint64, v int) bool {
		t.Log(k, "=>", v)
		count++
		return count < 5
	})
}

func TestSet_ScanReverse(t *testing.T) {
	var set = containers.NewSet[uint64, int](10)
	for i := range 10 {
		set.Push(uint64(i), rand.Int()/1_000_000)
	}
	var count int
	set.ScanReverse(func(k uint64, v int) bool {
		t.Log(k, "=>", v)
		count++
		return count < 5
	})
}

func TestSet_UpsertFunc(t *testing.T) {
	var a = assert.NewAssertion(t)

	var set = containers.NewSet[uint64, uint64](10)
	set.Push(1, 11)

	{
		set.UpsertFunc(1, func(value uint64) (resultValue uint64) {
			return value + 1
		})

		v, ok := set.Value(1)
		a.IsTrue(ok)
		a.IsTrue(v == 12)
	}
	set.Inspect(t)

	{
		set.UpsertFunc(1, func(value uint64) (resultValue uint64) {
			return value + 100
		})

		v, ok := set.Value(1)
		a.IsTrue(ok)
		a.IsTrue(v == 112)
	}
	set.Inspect(t)

	{
		set.UpsertFunc(2, func(value uint64) (resultValue uint64) {
			return 1000
		})
		v, ok := set.Value(2)
		a.IsTrue(ok)
		a.IsTrue(v == 1000)
	}
	set.Inspect(t)

	{
		set.UpsertFunc(2, func(value uint64) (resultValue uint64) {
			return 1000
		})
		v, ok := set.Value(2)
		a.IsTrue(ok)
		a.IsTrue(v == 1000)
	}
	set.Inspect(t)
}

func TestSet_Evict_Change_Keys(t *testing.T) {
	var a = assert.NewAssertion(t)

	var countEvicted = 0

	var set = containers.NewSet[uint64, uint64](10)
	set.OnEvict(func(keys []uint64) {
		t.Log("evicted:", keys)

		if slices.Contains(keys, 10_000) {
			t.Fatal("keys should not contain 10000")
		}

		countEvicted += len(keys)
		keys[0] = 10000
		time.Sleep(10 * time.Millisecond) // cost
	})

	for i := range 1_000 {
		set.Push(uint64(i), uint64(i))
	}

	time.Sleep(1 * time.Second)

	a.IsTrue(set.Len() == 10)
	a.IsTrue(countEvicted == 990)

	t.Log(set.Len(), "keys", countEvicted, "evicted")
}

func TestSet_Evict_Concurrent(t *testing.T) {
	var a = assert.NewAssertion(t)

	var countEvicted = 0

	var set = containers.NewSet[uint64, uint64](10)
	set.OnEvict(func(keys []uint64) {
		t.Log("evicted:", keys)
		countEvicted += len(keys)

		if slices.Contains(keys, 10_000) {
			t.Fatal("keys should not contain 10000")
		}

		go func() {
			keys[0] = 10_000
			time.Sleep(10 * time.Millisecond) // cost
		}()
	})

	for i := range 1_000 {
		set.Push(uint64(i), uint64(i))
	}

	time.Sleep(1 * time.Second)

	a.IsTrue(set.Len() == 10)
	a.IsTrue(countEvicted == 990)

	t.Log(set.Len(), "keys", countEvicted, "evicted")
}

func TestSet_EvictAll(t *testing.T) {
	var set = containers.NewSet[uint64, uint64](100).OnEvict(func(keys []uint64) {
		t.Log("evicted:", keys)
	})
	for i := range 20 {
		set.Push(uint64(i), uint64(i*100))
	}

	set.EvictAll(func(value uint64) bool {
		return true
	}, func(evictedKeys []uint64) {
		t.Log("evicted2:", evictedKeys)
	})

	time.Sleep(50 * time.Millisecond)

	set.Inspect(t)
}

func TestSet_Evict(t *testing.T) {
	var a = assert.NewAssertion(t)

	var set = containers.NewSet[uint64, uint64](100).OnEvict(func(keys []uint64) {
		t.Log("evicted:", keys)
	})
	for i := range 21 {
		set.Push(uint64(i), uint64(i*100))
	}
	a.IsTrue(set.Len() == 21)

	set.Evict(10, func(value uint64) bool {
		return true
	}, func(evictedKeys []uint64) {
		t.Log("evicted0:", evictedKeys)
	})

	time.Sleep(50 * time.Millisecond)

	set.Inspect(t)

	a.IsTrue(set.Len() == 11)
}

func TestSet_EvictKey(t *testing.T) {
	var set = containers.NewSet[uint64, uint64](100)
	set.OnEvict(func(keys []uint64) {
		t.Log("evicted:", keys)
	})

	set.Push(1, 1)
	set.Push(2, 2)
	set.Push(3, 3)
	set.EvictKey(2)
	time.Sleep(200 * time.Millisecond)

	set.Close()
	set.EvictKey(2)

	set.Close()
}

func TestSet_Evict_Loop(t *testing.T) {
	var a = assert.NewAssertion(t)

	var set = containers.NewSet[uint64, uint64](100).OnEvict(func(keys []uint64) {
		t.Log("evicted:", keys)
	})
	for i := range 21 {
		set.Push(uint64(i), uint64(i*100))
	}
	a.IsTrue(set.Len() == 21)

	for i := range 4 {
		set.Evict(10, func(value uint64) bool {
			return true
		}, func(evictedKeys []uint64) {
			t.Log("evicted"+types.String(i)+":", evictedKeys)
		})
	}

	time.Sleep(50 * time.Millisecond)

	set.Inspect(t)

	a.IsTrue(set.Len() == 0)
}

func TestSet_Evict_Over(t *testing.T) {
	var a = assert.NewAssertion(t)

	var set = containers.NewSet[uint64, uint64](100).OnEvict(func(keys []uint64) {
		t.Log("evicted:", keys)
	})
	for i := range 21 {
		set.Push(uint64(i), uint64(i*100))
	}

	for i := range 4 {
		set.Evict(100, func(value uint64) bool {
			return true
		}, func(evictedKeys []uint64) {
			t.Log("evicted"+types.String(i)+":", evictedKeys)
		})
	}

	set.Inspect(t)

	time.Sleep(50 * time.Millisecond)

	a.IsTrue(set.Len() == 0)
}

func TestSet_UniqueId(t *testing.T) {
	var m = map[int]int{}

	for range 100 {
		var set = containers.NewSet[uint64, uint64](5)
		t.Log(set.UniqueId())
		m[set.UniqueId()%runtime.NumCPU()]++
	}

	t.Log(m)
}

func TestSet_Close(t *testing.T) {
	var set = containers.NewSet[uint64, uint64](5)
	set.Push(1, 1)
	set.Push(2, 2)
	set.Push(3, 3)
	set.Push(4, 4)
	set.Push(5, 5)
	set.Close()

	time.Sleep(200 * time.Microsecond)

	set.Push(6, 6)
	set.Push(7, 7)
}

func BenchmarkSet_Push(b *testing.B) {
	var set = containers.NewSet[uint64, uint64](10_000)
	set.OnEvict(func(keys []uint64) {
		_ = keys
	})

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var r = rand.Uint64()
			set.Push(r, r%10_000)
		}
	})
}

func BenchmarkSet_Push_Many(b *testing.B) {
	var set = containers.NewSet[uint64, uint64](10_000)
	set.OnEvict(func(keys []uint64) {
		_ = keys
	})

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var r = rand.Uint64()
			set.Push(r, r%100_000)
		}
	})
}

func BenchmarkSet_Push_Pause(b *testing.B) {
	var set = containers.NewSet[uint64, uint64](10_000)
	set.OnEvict(func(keys []uint64) {
		time.Sleep(1 * time.Millisecond)
	})

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var r = rand.Uint64()
			set.Push(r, r)
		}
	})
}

func BenchmarkSet_Push_Concurrent(b *testing.B) {
	runtime.GOMAXPROCS(128)

	var set = containers.NewSet[uint64, uint64](10_000)
	set.OnEvict(func(keys []uint64) {
		if len(keys) > 0 && keys[0] == 1000000 {
			b.Fatal()
		}

		keys[0] = 1000000

		time.Sleep(300 * time.Microsecond)
	})

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var r = rand.Uint64()
			set.Push(r, r)
		}
	})
}

func BenchmarkSet_Evict(b *testing.B) {
	var set = containers.NewSet[uint64, uint64](10_000)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			set.Evict(100, func(value uint64) bool {
				return true
			}, func(evictedKeys []uint64) {
				_ = evictedKeys
			})
		}
	})
}
