package restcache

import (
	"context"
	"time"

	"github.com/wencan/fastrest/resterror"
	"github.com/wencan/gox/xsync/sentinel"
)

// Storage 缓存存储接口。一般是对接redis、lru，透明处理业务数据。
type Storage interface {
	// Get 查询存储的数据。
	// 实现逻辑应该处理掉需要忽略的错误。
	Get(ctx context.Context, key string, valuePtr interface{}) (found bool, err error)

	// Set 存储数据。
	// 实现逻辑应该处理掉需要忽略的错误。
	Set(ctx context.Context, key string, value interface{}, TTL time.Duration) error
}

// QueryFunc 查询未缓存的数据。一般是调http/rpc接口、查持久化数据库。
type QueryFunc func(ctx context.Context, destPtr interface{}, args interface{}) (found bool, err error)

// Caching 缓存。
type Caching struct {
	// Storage 存储接口。
	Storage Storage

	// Query 如果没从缓存里找到，调用Query查询。
	Query QueryFunc

	// TTLRange 缓存生存时间区间。每次随机取一个区间内的值。
	TTLRange [2]time.Duration

	// sentinelGroup  哨兵机制。
	sentinelGroup sentinel.SentinelGroup

	// SentinelTTL 哨兵和哨兵持有的临时缓存的生存时间。用来省去双重检查。一般1s即可。
	// “副作用”是可以避免高频查询找不到的数据。
	SentinelTTL time.Duration
}

// Get 查询。dest为对象地址，key为缓存key，args为查询函数参数。
// dest的值很可能是共享的，内容数据不可修改。
func (caching *Caching) Get(ctx context.Context, destPtr interface{}, key string, args interface{}) (found bool, err error) {
	// 先查缓存
	found, err = caching.Storage.Get(ctx, key, destPtr)
	if err != nil {
		// 错误
		// Storage的实现逻辑应该处理掉需要忽略的错误
		return false, err
	} else if found {
		// 没错误
		// 找到
		return true, nil
	}

	// 哨兵机制。同一进程内，同一时间，不同查询同key的数据
	err = caching.sentinelGroup.Do(ctx, destPtr, key, args, func(ctx context.Context, destPtr interface{}, args interface{}) error {
		found, err := caching.Query(ctx, destPtr, args)
		if err != nil {
			// Query的实现逻辑应该处理掉需要忽略的错误
			return err
		}
		if !found {
			return resterror.FormatNotFoundError("not found [%s]", key)
		}

		// 保存
		err = caching.Storage.Set(ctx, key, destPtr, getTTL(caching.TTLRange))
		if err != nil {
			// Storage的实现逻辑应该处理掉需要忽略的错误
			return err
		}

		return nil
	})
	// 延迟删除哨兵（和哨兵持有的临时缓存）
	// 省去双重检查。
	time.AfterFunc(caching.SentinelTTL, func() {
		caching.sentinelGroup.Delete(key)
	})

	if err != nil {
		if resterror.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
