// Copyright 2026 FlexCDN root@flexcdn.cn. All rights reserved. Official site: https://flexcdn.cn .

package containers

import (
	"math"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/tidwall/btree"
	"github.com/tjbrains/TeaGo/logs"
)

type NumberType interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | string
}

const NoLimit = math.MaxInt

type Set[K comparable, N NumberType] struct {
	rawMap *btree.Map[N, map[K]struct{}] // value => [ {key1: Zero}, {key2: Zero}, ... ]
	keyMap map[K]N                       // key => value

	capacity     int
	countDeletes uint64

	evictUniqueId int
	onEvict       func(keys []K)

	closer        chan bool
	isClosedValue atomic.Bool

	mu *sync.RWMutex
}

func NewSet[K comparable, N NumberType](capacity int) *Set[K, N] {
	if capacity <= 0 {
		capacity = 1_000_000
	}

	var set = &Set[K, N]{
		rawMap:        &btree.Map[N, map[K]struct{}]{},
		keyMap:        make(map[K]N),
		capacity:      capacity,
		mu:            &sync.RWMutex{},
		closer:        make(chan bool, 1),
		isClosedValue: atomic.Bool{},
		evictUniqueId: rand.IntN(1 << 10),
	}

	return set
}

func (this *Set[K, N]) SetCapacity(capacity int) {
	if capacity > 0 {
		this.capacity = capacity
	}
}

func (this *Set[K, N]) OnEvict(onEvict func(keys []K)) *Set[K, N] {
	this.onEvict = onEvict
	return this
}

func (this *Set[K, N]) EvictKey(key K) {
	if this.isClosed() {
		return
	}
	if this.onEvict != nil {
		pushEvict(this, key)
	}
}

func (this *Set[K, N]) Pop() (K, bool) {
	this.mu.Lock()
	popKey, popOk := this.pop()
	this.mu.Unlock()

	if popOk {
		this.EvictKey(popKey)
	}

	return popKey, popOk
}

func (this *Set[K, N]) TryPush(key K, n N) bool {
	return this.push(key, n, false)
}

func (this *Set[K, N]) Push(key K, n N) {
	this.push(key, n, true)
}

func (this *Set[K, N]) Delete(key ...K) {
	this.mu.Lock()

	for _, k := range key {
		value, ok := this.keyMap[k]
		if ok {
			subKeyMap, exists := this.rawMap.Get(value)
			if exists {
				delete(subKeyMap, k)
				if len(subKeyMap) == 0 {
					this.rawMap.Delete(value)
				}
			}
			delete(this.keyMap, k)
		}
	}

	this.mu.Unlock()

	// 这里不触发 onEvict
}

func (this *Set[K, N]) Value(key K) (v N, ok bool) {
	this.mu.RLock()
	defer this.mu.RUnlock()
	v, ok = this.keyMap[key]
	return
}

func (this *Set[K, N]) UpsertFunc(key K, updateFunc func(n N) (resultN N)) N {
	v, _ := this.upsertFunc(key, updateFunc, true)
	return v
}

func (this *Set[K, N]) TryUpsertFunc(key K, updateFunc func(n N) (resultValue N)) (N, bool) {
	return this.upsertFunc(key, updateFunc, false)
}

func (this *Set[K, N]) upsertFunc(key K, updateFunc func(value N) (resultValue N), force bool) (resultValue N, resultOk bool) {
	if force {
		this.mu.Lock()
	} else {
		if !this.mu.TryLock() {
			return
		}
	}

	oldValue, existOld := this.keyMap[key]
	var newValue = updateFunc(oldValue)

	if existOld && oldValue == newValue {
		this.mu.Unlock()
		return newValue, true
	}

	var popKey K
	var popOk bool

	// not exists, push one
	if !existOld {
		// check capacity
		if len(this.keyMap) >= this.capacity {
			popKey, popOk = this.pop()
		}
	}

	// update value
	this.keyMap[key] = newValue

	// old
	oldSubKeyMap, exists := this.rawMap.Get(oldValue)
	if exists {
		delete(oldSubKeyMap, key)
		if len(oldSubKeyMap) == 0 {
			this.rawMap.Delete(oldValue)
		}
	}

	// new
	newSubKeyMap, exists := this.rawMap.Get(newValue)
	if exists {
		newSubKeyMap[key] = struct{}{}
	} else {
		this.rawMap.Set(newValue, map[K]struct{}{key: {}})
	}

	this.mu.Unlock()

	if popOk && this.onEvict != nil {
		this.EvictKey(popKey)
	}

	return newValue, true
}

func (this *Set[K, N]) Contains(key K) bool {
	this.mu.RLock()
	defer this.mu.RUnlock()
	_, ok := this.keyMap[key]
	return ok
}

func (this *Set[K, N]) UpdateValue(key K, newValue N) {
	this.mu.Lock()
	defer this.mu.Unlock()

	oldValue, ok := this.keyMap[key]
	if !ok {
		return
	}
	if oldValue == newValue {
		return
	}

	this.keyMap[key] = newValue

	// old
	oldSubKeyMap, exists := this.rawMap.Get(oldValue)
	if exists {
		delete(oldSubKeyMap, key)
		if len(oldSubKeyMap) == 0 {
			this.rawMap.Delete(oldValue)
		}
	}

	// new
	newSubKeyMap, exists := this.rawMap.Get(newValue)
	if exists {
		newSubKeyMap[key] = struct{}{}
	} else {
		this.rawMap.Set(newValue, map[K]struct{}{key: {}})
	}
}

func (this *Set[K, N]) Len() int {
	this.mu.RLock()
	var l = len(this.keyMap)
	this.mu.RUnlock()
	return l
}

func (this *Set[K, N]) Capacity() int {
	return this.capacity
}

func (this *Set[K, N]) Keys() []K {
	this.mu.RLock()

	var keys []K

	this.rawMap.Scan(func(value N, v map[K]struct{}) bool {
		for key := range v {
			keys = append(keys, key)
		}
		return true
	})

	this.mu.RUnlock()
	return keys
}

func (this *Set[K, N]) Scan(iter func(k K, v N) bool) {
	this.mu.RLock()
	this.rawMap.Scan(func(n N, keyMap map[K]struct{}) bool {
		for key := range keyMap {
			if !iter(key, n) {
				return false
			}
		}
		return true
	})
	this.mu.RUnlock()
}

func (this *Set[K, N]) ScanReverse(iter func(k K, v N) bool) {
	this.mu.RLock()
	this.rawMap.Reverse(func(n N, keyMap map[K]struct{}) bool {
		for key := range keyMap {
			if !iter(key, n) {
				return false
			}
		}
		return true
	})
	this.mu.RUnlock()
}

func (this *Set[K, N]) EvictAll(shouldEvict func(n N) bool, onEvict func(evictedKeys []K)) {
	this.Evict(-1, shouldEvict, onEvict)
}

func (this *Set[K, N]) Evict(count int, shouldEvict func(n N) bool, onEvict func(evictedKeys []K)) int {
	if count == 0 {
		return 0
	}

	this.mu.Lock()

	if this.rawMap.Len() == 0 {
		this.mu.Unlock()
		return 0
	}

	var evictedKeys = make([]K, 0, 256)
	var evictedValues = make([]N, 0, 8)

	// lookup expired keys
	var noLimit = count < 0
	this.rawMap.Scan(func(n N, subKeyMap map[K]struct{}) bool {
		if shouldEvict == nil || shouldEvict(n) {
			if noLimit {
				for key := range subKeyMap {
					evictedKeys = append(evictedKeys, key)
					delete(this.keyMap, key)
				}
				evictedValues = append(evictedValues, n)
			} else {
				for key := range subKeyMap {
					evictedKeys = append(evictedKeys, key)
					delete(this.keyMap, key)
					delete(subKeyMap, key)

					count--
					if count <= 0 {
						break
					}
				}
				if len(subKeyMap) == 0 {
					evictedValues = append(evictedValues, n)
				}
				return count > 0
			}
			return true
		}
		return false
	})

	// remove value from map
	if len(evictedValues) > 0 {
		for _, value := range evictedValues {
			this.rawMap.Delete(value)
		}
	}

	if len(evictedKeys) > 0 {
		// free memory
		this.countDeletes += uint64(len(evictedKeys))
		this.freeMemory()
	}

	this.mu.Unlock()

	if len(evictedKeys) > 0 {
		if onEvict != nil {
			onEvict(evictedKeys)
		} else if this.onEvict != nil {
			this.onEvict(evictedKeys)
		}
	}

	return len(evictedKeys)
}

func (this *Set[K, N]) Clear() {
	this.mu.Lock()
	defer this.mu.Unlock()

	this.keyMap = make(map[K]N)
	this.rawMap.Clear()
}

func (this *Set[K, N]) Lock() {
	this.mu.Lock()
}

func (this *Set[K, N]) Unlock() {
	this.mu.Unlock()
}

func (this *Set[K, N]) RLock() {
	this.mu.RLock()
}

func (this *Set[K, N]) RUnlock() {
	this.mu.RUnlock()
}

func (this *Set[K, N]) UniqueId() int {
	return this.evictUniqueId
}

func (this *Set[K, N]) Close() {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.isClosed() {
		return
	}
	this.isClosedValue.Store(true)

	select {
	case this.closer <- true:
	default:
	}
}

func (this *Set[K, N]) Inspect(t *testing.T) {
	t.Log("===[KEY => NUMBER]===")
	logs.PrintAsJSON(this.keyMap, t)

	t.Log("===[NUMBER => KEYS]===")
	var valueKeysMap = map[N][]K{}
	this.rawMap.Scan(func(value N, keys map[K]struct{}) bool {
		var keyList []K
		for key := range keys {
			keyList = append(keyList, key)
		}
		valueKeysMap[value] = keyList
		return true
	})
	logs.PrintAsJSON(valueKeysMap, t)
}

func (this *Set[K, N]) push(key K, newValue N, force bool) bool {
	var popKey K
	var popOk bool

	if force {
		this.mu.Lock()
	} else {
		if !this.mu.TryLock() {
			return false
		}
	}

	// delete old
	oldValue, ok := this.keyMap[key]
	if ok {
		if newValue == oldValue {
			this.mu.Unlock()
			return true
		}

		subKeyMap, exists := this.rawMap.Get(oldValue)
		if exists {
			delete(subKeyMap, key)
			if len(subKeyMap) == 0 {
				this.rawMap.Delete(oldValue)
			}
		}
	} else {
		// check capacity
		if len(this.keyMap) >= this.capacity {
			popKey, popOk = this.pop()
		}
	}

	this.keyMap[key] = newValue

	subKeyMap, exists := this.rawMap.Get(newValue)
	if exists {
		subKeyMap[key] = struct{}{}
	} else {
		this.rawMap.Set(newValue, map[K]struct{}{
			key: {},
		})
	}

	this.mu.Unlock()

	if popOk {
		// 这里不能直接调用onEvict，否则可能会导致外部调用的死锁
		if this.onEvict != nil && !this.isClosed() {
			this.EvictKey(popKey)
		}
	}

	return true
}

func (this *Set[K, N]) pop() (key K, ok bool) {
	minValue, minSubKeyMap, exists := this.rawMap.Min()
	if !exists {
		return
	}

	for subKey := range minSubKeyMap {
		key = subKey
		ok = true

		delete(minSubKeyMap, subKey)
		delete(this.keyMap, subKey)

		if len(minSubKeyMap) == 0 {
			this.rawMap.Delete(minValue)
		}

		break
	}

	return
}

func (this *Set[K, N]) evict(keys []any) {
	if this.isClosed() {
		return
	}

	if this.onEvict != nil {
		var newKeys = make([]K, 0, len(keys))
		for _, key := range keys {
			newKeys = append(newKeys, key.(K))
		}

		this.onEvict(newKeys)
	}
}

func (this *Set[K, N]) freeMemory() {
	if this.countDeletes < 100_000_000 || len(this.keyMap) > 2_000 {
		return
	}

	this.countDeletes = 0

	// key map
	var newKeyMap = make(map[K]N, len(this.keyMap))
	for k, v := range this.keyMap {
		newKeyMap[k] = v
	}
	this.keyMap = newKeyMap
}

func (this *Set[K, N]) isClosed() bool {
	return this.isClosedValue.Load()
}
