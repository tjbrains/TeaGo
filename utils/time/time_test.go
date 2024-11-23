package timeutil_test

import (
	timeutil "github.com/tjbrains/TeaGo/utils/time"
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	t.Log("Y-m-d:", timeutil.Format("Y-m-d"))
	t.Log("Ymd:", timeutil.Format("Ymd"))
	t.Log("Ym:", timeutil.Format("Ym"))
	t.Log("Y-m-d H:i:s", timeutil.Format("Y-m-d H:i:s"))
	t.Log("Y/m/d H:i:s", timeutil.Format("Y/m/d H:i:s"))
	t.Log("Hi:", timeutil.Format("Hi"))
	t.Log("His:", timeutil.Format("His"))
	t.Log(timeutil.Format("Y-m-d H:i:s", time.Date(2020, 10, 10, 0, 0, 0, 0, time.Local)))
	t.Log(timeutil.Format("c", time.Now().Add(-1*time.Hour)))
	t.Log(timeutil.Format("r"))
	t.Log(timeutil.Format("U"))
	t.Log(timeutil.Format("D"))
	t.Log(timeutil.Format("l"))
	t.Log(timeutil.Format("A"))
	t.Log(timeutil.Format("a"))
	t.Log(timeutil.Format("F"))
	t.Log(timeutil.Format("Y, y"))
	t.Log(timeutil.Format("g, h"))
	t.Log(timeutil.Format("u, v"))
	t.Log(timeutil.Format("O, P"))
}

func BenchmarkFormat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		timeutil.Format("Y-m-d H:i:s")
	}
}
