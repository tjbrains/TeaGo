// Copyright 2026 FlexCDN root@flexcdn.cn. All rights reserved. Official site: https://flexcdn.cn .

package dbs_test

import (
	"strconv"
	"testing"

	"github.com/tjbrains/TeaGo/assert"
	"github.com/tjbrains/TeaGo/dbs"
)

func TestFieldValues(t *testing.T) {
	var a = assert.NewAssertion(t)

	var v = dbs.NewFieldValues[any]()
	a.IsTrue(v.Len() == 0)

	v.Set("name", "Lily")
	a.IsTrue(v.Len() == 1)
	v.Set("name", "lucy")
	a.IsTrue(v.Len() == 1)

	value, ok := v.Get("name")
	a.IsTrue(ok)
	a.IsTrue(value == "lucy")

	v.Set("age", 20)
	v.Set("gender", 1)
	v.Set("book", "golang")

	t.Log("=== before sorted ===")
	for field, value := range v.Iterator() {
		t.Log(field, "=>", value)
	}

	t.Log("=== sorted ===")
	v.SortKeys()
	for field, value := range v.Iterator() {
		t.Log(field, "=>", value)
	}
}

func TestFieldValues_Append(t *testing.T) {
	var a = assert.NewAssertion(t)

	var v = dbs.NewFieldValues[any]()
	v.Append("name", "Lily")
	v.Append("name", "Lucy")
	a.IsTrue(v.Len() == 2)
	for field, value := range v.Iterator() {
		t.Log(field, "=>", value)
	}

	t.Log(v.ToMap())
}

func BenchmarkFieldValues(b *testing.B) {
	var v = dbs.NewFieldValues[int]()

	b.ReportAllocs()
	b.ResetTimer()

	for i := range b.N {
		var r = i % 10
		v.Set(strconv.Itoa(r), r)
	}
}
