package lrucache

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/wencan/fastrest/restutils"
	"github.com/wencan/gox/xsync"
)

type lruEntry struct {
	value interface{}

	expire time.Time
}

var lruEntryPool = sync.Pool{New: func() interface{} {
	return &lruEntry{}
}}

// LRUCache 进程内的LRU缓存。只存储最近使用的。
// 基于github.com/wencan/xsync/LRUMap，按批次（区块）清理最近不用的数据。
// 实现了github.com/wencan/fastrest/restcache的Storage接口和MStorage接口。
type LRUCache struct {
	lruMap *xsync.LRUMap
}

// NewLRUCache 创建lru缓存。chunkCapacity、chunkNum是github.com/wencan/xsync/LRUMap的区块参数。
func NewLRUCache(chunkCapacity int, chunkNum int) *LRUCache {
	return &LRUCache{
		// LRUMap的清理逻辑有bug。暂时不用清理
		// lruMap: xsync.NewLRUMapWithEvict(chunkCapacity, chunkNum, func(key, value interface{}) {
		// 	lruEntryPool.Put(value)
		// }),
		lruMap: xsync.NewLRUMap(chunkCapacity, chunkNum),
	}
}

// Get 实现github.com/wencan/fastrest/restcache的Storage接口。
func (lru *LRUCache) Get(ctx context.Context, key string, valuePtr interface{}) (found bool, err error) {
	value, upgrade, ok := lru.lruMap.SilentLoad(key)
	if !ok {
		return false, nil
	}

	entry := value.(*lruEntry)
	// if time.Now().After(entry.expire) {
	if restutils.CoarseTimestamp() > float64(entry.expire.UnixMilli())/1000 {
		// 过期
		return false, nil
	}

	// 赋值
	v := reflect.ValueOf(entry.value)
	reflect.ValueOf(valuePtr).Elem().Set(v)

	// 更新为最近使用
	upgrade()

	return true, nil
}

// Set 实现github.com/wencan/fastrest/restcache的Storage接口。
func (lru *LRUCache) Set(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
	entry := lruEntryPool.Get().(*lruEntry)
	entry.value = value
	entry.expire = time.Now().Add(TTL)

	lru.lruMap.Store(key, entry)
	return nil
}

// MGet 实现github.com/wencan/fastrest/restcache的MStorage接口。
func (lru *LRUCache) MGet(ctx context.Context, keys []string, destSlicePtr interface{}) (missIndexes []int, err error) {
	destSliceValue := reflect.ValueOf(destSlicePtr).Elem()

	for index, key := range keys {
		value, upgrade, ok := lru.lruMap.SilentLoad(key)
		if !ok {
			missIndexes = append(missIndexes, index)
			continue
		}

		entry := value.(*lruEntry)
		// if time.Now().After(entry.expire) {
		if restutils.CoarseTimestamp() > float64(entry.expire.UnixMilli())/1000 {
			// 过期
			missIndexes = append(missIndexes, index)
			continue
		}

		// 赋值
		v := reflect.ValueOf(entry.value)
		destSliceValue.Set(reflect.Append(destSliceValue, v))

		// 更新为最近使用
		upgrade()
	}

	return missIndexes, nil
}

// MSet 实现github.com/wencan/fastrest/restcache的MStorage接口。
func (lru *LRUCache) MSet(ctx context.Context, keys []string, destSlice interface{}, ttl time.Duration) error {
	destSliceValue := reflect.ValueOf(destSlice)
	if len(keys) != destSliceValue.Len() {
		return errors.New("wrong arguments")
	}

	for index, key := range keys {
		lru.Set(ctx, key, destSliceValue.Index(index).Interface(), ttl)
	}

	return nil
}
