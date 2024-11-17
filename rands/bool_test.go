// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package rands_test

import (
	"github.com/tjbrains/TeaGo/rands"
	"testing"
)

func TestRand_Bool_Distribute_1(t *testing.T) {
	var m = map[bool]int{} // number => count
	for i := 0; i < 1000000; i++ {
		var v = rands.Bool()
		_, ok := m[v]
		if ok {
			m[v]++
		} else {
			m[v] = 1
		}
	}
	t.Log(m)
}
