package caches

import "time"

// Item 缓存条目定义
type Item struct {
	Key        string
	Value      interface{}
	expireTime time.Time
}

// Set 设置值
func (this *Item) Set(value interface{}) *Item {
	this.Value = value
	return this
}

// ExpireAt 设置过期时间
func (this *Item) ExpireAt(expireTime time.Time) *Item {
	this.expireTime = expireTime
	return this
}

// Expire 设置过期时长
func (this *Item) Expire(duration time.Duration) *Item {
	return this.ExpireAt(time.Now().Add(duration))
}

// IsExpired 判断是否已过期
func (this *Item) IsExpired() bool {
	return time.Since(this.expireTime) > 0
}
