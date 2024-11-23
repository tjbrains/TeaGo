package maps

import (
	"cmp"
	"github.com/tjbrains/TeaGo/lists"
	"github.com/tjbrains/TeaGo/types"
	"slices"
)

type OrderedMap[K cmp.Ordered, V any] struct {
	keys      []K
	valuesMap map[K]V
}

func NewOrderedMap[K cmp.Ordered, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		valuesMap: map[K]V{},
	}
}

// Keys 取得所有Key
func (this *OrderedMap[K, V]) Keys() []K {
	return this.keys
}

// Sort 根据元素值进行排序
func (this *OrderedMap[K, V]) Sort() {
	lists.Sort(this.keys, func(i int, j int) bool {
		var value1 = this.valuesMap[this.keys[i]]
		var value2 = this.valuesMap[this.keys[j]]

		return types.Compare(value1, value2) <= 0
	})
}

// SortKeys 根据Key进行排序
func (this *OrderedMap[K, V]) SortKeys() {
	slices.Sort[[]K, K](this.keys)
}

// Reverse 翻转键
func (this *OrderedMap[K, V]) Reverse() {
	slices.Reverse(this.keys)
}

// Put 添加元素
func (this *OrderedMap[K, V]) Put(key K, value V) {
	_, ok := this.valuesMap[key]
	if !ok {
		this.keys = append(this.keys, key)
	}
	this.valuesMap[key] = value
}

// Get 取得元素值
func (this *OrderedMap[K, V]) Get(key K) (value V, ok bool) {
	value, ok = this.valuesMap[key]
	return
}

// Delete 删除元素
func (this *OrderedMap[K, V]) Delete(key K) {
	var index = -1
	for itemIndex, itemKey := range this.keys {
		if itemKey == key {
			index = itemIndex
			break
		}
	}
	if index > -1 {
		this.keys = append(this.keys[0:index], this.keys[index+1:]...)
		delete(this.valuesMap, key)
	}
}

// Range 对每个元素执行迭代器
func (this *OrderedMap[K, V]) Range(iterator func(key K, value V)) {
	for _, key := range this.keys {
		iterator(key, this.valuesMap[key])
	}
}

// Len 取得Map的长度
func (this *OrderedMap[K, V]) Len() int {
	return len(this.keys)
}
