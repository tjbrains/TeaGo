package rands

import (
	"math/rand/v2"
)

// Int 随机获取一个Int数字
func Int(min int, max int) int {
	if min > max {
		min, max = max, min
	}
	var r = max - min + 1
	if r == 0 {
		return min
	}

	return min + (rand.Int() % r)
}

// Int64 随机获取一个Int64数字
func Int64() int64 {
	return rand.Int64()
}
