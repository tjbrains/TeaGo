// Copyright 2026 FlexCDN root@flexcdn.cn. All rights reserved. Official site: https://flexcdn.cn .

package containers

import (
	"testing"
	"time"
)

type LRUFunc[K comparable] func(lru *LRU[K])

func LRURoundTouch[K comparable]() LRUFunc[K] {
	return func(lru *LRU[K]) {
		lru.roundTouch = true
	}
}

type LRU[K comparable] struct {
	rawSet *Set[K, int64]

	roundTouch bool
}

func NewLRU[K comparable](capacity int, opt ...LRUFunc[K]) *LRU[K] {
	var lru = &LRU[K]{
		rawSet: NewSet[K, int64](capacity),
	}

	for _, o := range opt {
		o(lru)
	}

	return lru
}

func (this *LRU[K]) SetCapacity(capacity int) {
	if capacity > 0 {
		this.rawSet.SetCapacity(capacity)
	}
}

func (this *LRU[K]) Capacity() int {
	return this.rawSet.Capacity()
}

func (this *LRU[K]) OnEvict(onEvict func(keys []K)) *LRU[K] {
	this.rawSet.OnEvict(onEvict)
	return this
}

func (this *LRU[K]) Touch(key K) {
	// upsert
	this.rawSet.UpsertFunc(key, func(value int64) (resultValue int64) {
		return this.nextId()
	})
}

func (this *LRU[K]) TouchN(key K, n int64) {
	// upsert
	this.rawSet.UpsertFunc(key, func(value int64) (resultValue int64) {
		return n
	})
}

func (this *LRU[K]) TryTouch(key K) {
	// upsert
	this.rawSet.TryUpsertFunc(key, func(value int64) (resultValue int64) {
		return this.nextId()
	})
}

func (this *LRU[K]) Contains(key K) bool {
	return this.rawSet.Contains(key)
}

func (this *LRU[K]) Delete(key ...K) {
	this.rawSet.Delete(key...)
}

func (this *LRU[K]) Evict(count int) int {
	return this.rawSet.Evict(count, func(value int64) bool {
		return true
	}, nil)
}

func (this *LRU[K]) EvictAll(shouldEvict func(n int64) bool, onEvict func(evictedKeys []K)) {
	this.rawSet.EvictAll(shouldEvict, onEvict)
}

func (this *LRU[K]) Keys() []K {
	return this.rawSet.Keys()
}

func (this *LRU[K]) N(key K) (int64, bool) {
	return this.rawSet.Value(key)
}

func (this *LRU[K]) Len() int {
	return this.rawSet.Len()
}

func (this *LRU[K]) Clear() {
	this.rawSet.Clear()
}

func (this *LRU[K]) Close() {
	this.rawSet.Close()
}

func (this *LRU[K]) Inspect(t *testing.T) {
	this.rawSet.Inspect(t)
}

func (this *LRU[K]) nextId() int64 {
	if this.roundTouch {
		return (time.Now().Unix()/10 + 1) * 10
	}
	return time.Now().Unix()
}
