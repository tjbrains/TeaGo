package actions

import (
	"github.com/tjbrains/TeaGo/maps"
	"sync"
)

// ActionContext 上下文变量容器
type ActionContext struct {
	context maps.Map
	locker  sync.RWMutex
}

// NewActionContext 获取新对象
func NewActionContext() *ActionContext {
	return &ActionContext{
		context: maps.Map{},
	}
}

// Set 设置变量
func (this *ActionContext) Set(key string, value interface{}) {
	this.locker.Lock()
	defer this.locker.Unlock()
	this.context[key] = value
}

// Get 获取变量
func (this *ActionContext) Get(key string) interface{} {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.context.Get(key)
}

// GetString 获取string变量
func (this *ActionContext) GetString(key string) string {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.context.GetString(key)
}

// GetInt 获取int变量
func (this *ActionContext) GetInt(key string) int {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.context.GetInt(key)
}

// GetInt64 获取int64变量
func (this *ActionContext) GetInt64(key string) int64 {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.context.GetInt64(key)
}

// GetBool 获取bool变量
func (this *ActionContext) GetBool(key string) bool {
	this.locker.RLock()
	defer this.locker.RUnlock()

	return this.context.GetBool(key)
}
