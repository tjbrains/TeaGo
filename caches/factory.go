package caches

import (
	"github.com/tjbrains/TeaGo/timers"
	"sync"
	"time"
)

// 操作类型
type CacheOperation = string

const (
	CacheOperationSet    = "set"
	CacheOperationDelete = "delete"
)

// Factory 缓存管理器
type Factory struct {
	items       map[string]*Item
	maxSize     int64                               // @TODO 实现maxSize
	onOperation func(op CacheOperation, item *Item) // 操作回调
	locker      *sync.Mutex
	looper      *timers.Looper
}

// NewFactory 创建一个新的缓存管理器
func NewFactory() *Factory {
	return NewFactoryInterval(30 * time.Second)
}

func NewFactoryInterval(duration time.Duration) *Factory {
	factory := &Factory{
		items:  map[string]*Item{},
		locker: &sync.Mutex{},
	}

	factory.looper = timers.Loop(duration, func(looper *timers.Looper) {
		factory.Clean()
	})

	return factory
}

// Set 设置缓存
func (this *Factory) Set(key string, value any, duration ...time.Duration) *Item {
	item := new(Item)
	item.Key = key
	item.Value = value

	if len(duration) > 0 {
		item.expireTime = time.Now().Add(duration[0])
	} else {
		item.expireTime = time.Now().Add(3600 * time.Second)
	}

	this.locker.Lock()
	if this.onOperation != nil {
		_, ok := this.items[key]
		if ok {
			this.onOperation(CacheOperationDelete, item)
		}
	}
	this.items[key] = item
	this.locker.Unlock()

	if this.onOperation != nil {
		this.onOperation(CacheOperationSet, item)
	}

	return item
}

// Get 获取缓存
func (this *Factory) Get(key string) (value any, found bool) {
	this.locker.Lock()
	defer this.locker.Unlock()

	item, found := this.items[key]
	if !found {
		return nil, false
	}

	if item.IsExpired() {
		return nil, false
	}

	return item.Value, true
}

// Has 判断是否有缓存
func (this *Factory) Has(key string) bool {
	_, found := this.Get(key)
	return found
}

// Delete 删除缓存
func (this *Factory) Delete(key string) {

	this.locker.Lock()
	item, ok := this.items[key]
	if ok {
		delete(this.items, key)
		if this.onOperation != nil {
			this.onOperation(CacheOperationDelete, item)
		}
	}
	this.locker.Unlock()
}

// OnOperation 设置操作回调
func (this *Factory) OnOperation(f func(op CacheOperation, item *Item)) {
	this.onOperation = f
}

// Close 关闭
func (this *Factory) Close() {
	this.locker.Lock()
	defer this.locker.Unlock()

	if this.looper != nil {
		this.looper.Stop()
		this.looper = nil
	}

	this.items = map[string]*Item{}
}

// Reset 重置状态
func (this *Factory) Reset() {
	this.locker.Lock()
	defer this.locker.Unlock()
	this.items = map[string]*Item{}
}

// Items 读取所有条目
func (this *Factory) Items() map[string]*Item {
	this.locker.Lock()
	defer this.locker.Unlock()
	return this.items
}

// Clean 清理过期的缓存
func (this *Factory) Clean() {
	this.locker.Lock()
	defer this.locker.Unlock()

	for _, item := range this.items {
		if item.IsExpired() {
			delete(this.items, item.Key)

			if this.onOperation != nil {
				this.onOperation(CacheOperationDelete, item)
			}
		}
	}
}
