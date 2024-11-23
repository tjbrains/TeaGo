package caches

import (
	"sync"
)

// CacheFactory 缓存集成器，用户让别的对象拥有使用缓存的能力
type CacheFactory struct {
	factory *Factory
	locker  sync.Once
}

// Cache 取得缓存管理器
func (this *CacheFactory) Cache() *Factory {
	this.locker.Do(func() {
		this.factory = NewFactory()
	})
	return this.factory
}
