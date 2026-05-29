// Copyright 2026 FlexCDN root@flexcdn.cn. All rights reserved. Official site: https://flexcdn.cn .

package dbs

import "sync"

// QueryPool 查询池
type QueryPool struct {
	pool *sync.Pool
}

func NewQueryPool() *QueryPool {
	return &QueryPool{
		pool: &sync.Pool{
			New: func() any {
				return &Query{}
			},
		},
	}
}

// Get 获取查询对象
func (this *QueryPool) Get(model any) *Query {
	var query = this.pool.Get().(*Query)
	query.Init(model)
	return query
}

// Put 回收查询对象
func (this *QueryPool) Put(query *Query) {
	this.pool.Put(query)
}
