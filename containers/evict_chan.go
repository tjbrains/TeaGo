// Copyright 2026 FlexCDN root@flexcdn.cn. All rights reserved. Official site: https://flexcdn.cn .

package containers

import (
	"runtime"
)

type EvictInterface interface {
	evict(keys []any)
	UniqueId() int
}

type EvictItem struct {
	set EvictInterface
	key any
}

var sharedEvictChanList []chan EvictItem
var numCPU = min(max(runtime.NumCPU()/4, 2), 8)
var testSingleEvictChan = false

const evictBatchSize = 32

func init() {
	const chanSize = 32 << 10

	for range numCPU {
		var c = make(chan EvictItem, chanSize)
		sharedEvictChanList = append(sharedEvictChanList, c)

		go func() {
			var keys = make([]any, 0, evictBatchSize)
			var lastSet EvictInterface

			for item := range c {
				if lastSet == item.set {
					keys = append(keys, item.key)
					if len(keys) >= evictBatchSize || len(c) == 0 /* no more left */ {
						lastSet.evict(keys)
						keys = keys[:0]
					}
				} else {
					// last
					if lastSet != nil && len(keys) > 0 {
						lastSet.evict(keys)
					}

					// current
					lastSet = item.set
					keys = keys[:0]
					keys = append(keys, item.key)

					if len(c) == 0 /* no more left */ {
						lastSet.evict(keys)
						keys = keys[:0]
					}
				}
			}
		}()
	}
}

func pushEvict(set EvictInterface, key any) {
	var c = sharedEvictChanList[set.UniqueId()%numCPU]
	if testSingleEvictChan {
		c = sharedEvictChanList[0]
	}
	select {
	case c <- EvictItem{set: set, key: key}:
	default:
		go func() {
			c <- EvictItem{set: set, key: key}
		}()
	}
}

func TestSingleEvictChan() {
	testSingleEvictChan = true
}
