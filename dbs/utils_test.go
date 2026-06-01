// Copyright 2026 FlexCDN root@flexcdn.cn. All rights reserved. Official site: https://flexcdn.cn .

package dbs_test

import (
	"math/rand/v2"
	"testing"

	"github.com/tjbrains/TeaGo/dbs"
)

func BenchmarkMakeSlice(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		var r = rand.IntN(10)
		_ = dbs.MakeSlice[string](r)
	}
}

func BenchmarkMakeSlice_Raw(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		var r = rand.IntN(10)
		_ = make([]string, 0, r)
	}
}
