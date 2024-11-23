package types_test

import (
	"fmt"
	"github.com/tjbrains/TeaGo/types"
	"math"
	"reflect"
	"runtime"
	"testing"
)

func TestConvert(t *testing.T) {
	t.Log(types.Int(123))
	t.Log(types.Int("123.456"))
	t.Log(types.Bool("abc"), types.Bool(123), types.Bool(false), types.Bool(true))
	t.Log(types.Float32("123.456"))
	t.Log(types.Compare("abc", "123"), types.Compare(123, "12.3"))
	t.Log(types.Byte(123), types.Byte(255))
	t.Log(types.String(1), types.String(int64(1024)), types.String(true), types.String("Hello, World"), types.String([]string{"Hello"}))

	result, err := types.Slice([]string{"1", "2", "3"}, reflect.TypeOf([]int64{}))
	if err != nil {
		t.Log("fail to convert slice")
	} else {
		t.Logf("%#v", result)
	}
}

func TestCompare(t *testing.T) {
	assert(t, types.Compare(1, 2) < 0)
	assert(t, types.Compare(3, 2) > 0)
	assert(t, types.Compare(2, 2) == 0)
	assert(t, types.Compare(12.345, "12.345") == 0)
	assert(t, types.Compare(12.345, "12.39") < 0)
	assert(t, types.Compare("a", "b") < 0)
	assert(t, types.Compare("Abc", "abc") < 0)
	assert(t, types.Compare("abc", "abc") == 0)
}

func TestIntN(t *testing.T) {
	assert(t, types.Int8("1") == 1)
	assert(t, types.Int8("1024") == math.MaxInt8)
	assert(t, types.Int8("-1024") == math.MinInt8)

	assert(t, types.Int16("1") == 1)
	assert(t, types.Int16(1024) == 1024)
	assert(t, types.Int16(-1024) == -1024)
	assert(t, types.Int16(123456789101112) == math.MaxInt16)

	assert(t, types.Int32("1") == 1)
	assert(t, types.Int32(1024) == 1024)
	assert(t, types.Int32(-1024) == -1024)
	assert(t, types.Int32(123456789101112) == math.MaxInt32)
	t.Log("maxInt32:", math.MaxInt32)

	{

		type A int32
		var a A = 1234
		assert(t, types.Int32(a) == 1234)
	}
	{
		type A float32
		var a A = 123.456
		assert(t, types.Int32(a) == 123)
	}

	assert(t, types.Int64("1") == 1)
	assert(t, types.Int64(1024) == 1024)
	assert(t, types.Int64(-1024) == -1024)
	assert(t, types.Int64(9223372036854775807) == math.MaxInt64)
	t.Log("maxInt64:", math.MaxInt64)

	assert(t, types.Uint8(123) == 123)
	assert(t, types.Uint8(1024) == math.MaxUint8)
	t.Log("maxUint8:", math.MaxUint8)

	assert(t, types.Uint16(123) == 123)
	assert(t, types.Uint16(65536) == math.MaxUint16)
	t.Log("maxUint16:", math.MaxUint16)

	assert(t, types.Uint64(123) == 123)
}

func TestString(t *testing.T) {
	t.Log(types.String(123))
	t.Log(types.String(123456))
	t.Log(types.String(123456.123456))
	t.Log(types.String(123456890123456))
	t.Log(types.String(float64(12345)))
	t.Log(types.String(float32(12345)))
	t.Log(types.String(12345.12345))
}

func TestIsSlice(t *testing.T) {
	assert(t, !types.IsSlice(nil))

	{
		var s []string = nil
		assert(t, types.IsSlice(s))
	}

	{
		var s interface{} = nil
		assert(t, !types.IsSlice(s))
	}

	{
		var s *[]string = nil
		assert(t, !types.IsSlice(s))
	}

	{
		assert(t, types.IsSlice([]string{"a", "b", "c"}))
	}
}

func TestIsMap(t *testing.T) {
	assert(t, !types.IsMap(nil))

	{
		var s map[string]interface{} = nil
		assert(t, types.IsMap(s))
	}

	{
		var s interface{} = nil
		assert(t, !types.IsMap(s))
	}

	{
		assert(t, types.IsMap(map[string]interface{}{
			"a": "b",
		}))
	}
}

func assert(t *testing.T, b bool) {
	if !b {
		t.Fail()
	}
}

func BenchmarkInt_String(b *testing.B) {
	runtime.GOMAXPROCS(1)

	for i := 0; i < b.N; i++ {
		_ = types.String(1024)
	}
}

func BenchmarkInt_Sprintf(b *testing.B) {
	runtime.GOMAXPROCS(1)

	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%d", 1024)
	}
}
