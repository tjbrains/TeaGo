package numberutil_test

import (
	numberutil "github.com/tjbrains/TeaGo/utils/number"
	"testing"
)

func TestRangeInt(t *testing.T) {
	t.Log(numberutil.RangeInt(30, 10, 4))
	t.Log(numberutil.RangeInt(10, 30, 4))
}

func TestTimes(t *testing.T) {
	numberutil.Times(10, func(i uint) {
		t.Log(i)
	})
}

func TestMaxInt64(t *testing.T) {
	t.Log(numberutil.MaxInt64())
	t.Log(numberutil.MaxInt64(1))
	t.Log(numberutil.MaxInt64(1, 2, 3, 4, 5))
}

func TestMinInt64(t *testing.T) {
	t.Log(numberutil.MinInt64())
	t.Log(numberutil.MinInt64(1))
	t.Log(numberutil.MinInt64(1, 2, 3, 4, 5))
}
