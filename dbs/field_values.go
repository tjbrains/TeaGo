// Copyright 2026 FlexCDN root@flexcdn.cn. All rights reserved. Official site: https://flexcdn.cn .

package dbs

import (
	"iter"
	"slices"
)

type FieldValue[V any] struct {
	Field string
	Value V
}

type FieldValues[V any] struct {
	values []FieldValue[V]
}

func NewFieldValues[V any]() *FieldValues[V] {
	return &FieldValues[V]{}
}

func (this *FieldValues[V]) Set(field string, value V) {
	for index, v := range this.values {
		if v.Field == field {
			v.Value = value
			this.values[index] = v
			return
		}
	}

	this.values = append(this.values, FieldValue[V]{
		Field: field,
		Value: value,
	})
}

// Append 附加一个字段值
//
// 直接加入，不检查是否重复，你需要自己确保字段不会重复
func (this *FieldValues[V]) Append(field string, value V) {
	this.values = append(this.values, FieldValue[V]{
		Field: field,
		Value: value,
	})
}

func (this *FieldValues[V]) Get(field string) (value V, ok bool) {
	for _, v := range this.values {
		if v.Field == field {
			return v.Value, true
		}
	}
	return
}

func (this *FieldValues[V]) SortKeys() {
	slices.SortFunc(this.values, func(v1 FieldValue[V], v2 FieldValue[V]) int {
		if v1.Field < v2.Field {
			return -1
		}
		return 1
	})
}

func (this *FieldValues[V]) Iterator() iter.Seq2[string, V] {
	return iter.Seq2[string, V](func(yield func(field string, value V) bool) {
		for _, v := range this.values {
			if !yield(v.Field, v.Value) {
				return
			}
		}
	})
}

func (this *FieldValues[V]) Reset() {
	this.values = this.values[:0]
}

func (this *FieldValues[V]) Len() int {
	return len(this.values)
}

func (this *FieldValues[V]) ToMap() map[string]V {
	var m = make(map[string]V, len(this.values))
	for _, v := range this.values {
		m[v.Field] = v.Value
	}
	return m
}
