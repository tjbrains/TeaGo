package stringutil_test

import (
	stringutil "github.com/tjbrains/TeaGo/utils/string"
	"log"
	"sync"
	"testing"
)

func TestRandString(t *testing.T) {
	var s = stringutil.Rand(10)
	t.Log(s, len(s))
}

func TestRandStringUnique(t *testing.T) {
	var m = map[string]bool{}
	var mu = sync.Mutex{}
	var wg = sync.WaitGroup{}
	wg.Add(10000)
	for i := 0; i < 10000; i++ {
		go func() {
			defer wg.Done()
			s := stringutil.Rand(16)
			mu.Lock()
			_, found := m[s]
			mu.Unlock()
			if found {
				log.Println("duplicated", s)
				return
			}
			mu.Lock()
			m[s] = true
			mu.Unlock()
		}()
	}
	wg.Wait()
	t.Log("all unique")
	t.Log(m)
}

func TestConvertID(t *testing.T) {
	t.Log(stringutil.ConvertID(1234567890))
}

func TestVersionCompare(t *testing.T) {
	t.Log(stringutil.VersionCompare("1.0", "1.0.3"))
	t.Log(stringutil.VersionCompare("2.0.3", "2.0.3"))
	t.Log(stringutil.VersionCompare("2", "2.1"))
	t.Log(stringutil.VersionCompare("1.1.2", "1.2.1"))
	t.Log(stringutil.VersionCompare("1.10.2", "1.2.1"))
	t.Log(stringutil.VersionCompare("1.14.2", "1.1234567.1"))
}

func TestParseFileSize(t *testing.T) {
	{
		s, _ := stringutil.ParseFileSize("1k")
		t.Logf("%f", s)
	}
	{
		s, _ := stringutil.ParseFileSize("1m")
		t.Logf("%f", s)
	}
	{
		s, _ := stringutil.ParseFileSize("1g")
		t.Logf("%f", s)
	}
}

func TestRegexpCompile(t *testing.T) {
	for range 3 {
		reg, err := stringutil.RegexpCompile(`^\d+$`)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(reg.MatchString("123"))
	}
}

func TestMD5(t *testing.T) {
	t.Log(stringutil.MD5("123456")) // e10adc3949ba59abbe56e057f20f883e
	t.Log(stringutil.MD5("123456"))
	t.Log(stringutil.MD5("123456"))
	t.Log(stringutil.MD5("123")) // 202cb962ac59075b964b07152d234b70
}

func BenchmarkMD5(b *testing.B) {
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var sum = stringutil.MD5("123456")
			if sum != "e10adc3949ba59abbe56e057f20f883e" {
				b.Fatal("fail:", sum)
			}
		}
	})
}

func BenchmarkRegexpCompile(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := stringutil.RegexpCompile(`^\d+$`)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
