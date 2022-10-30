package restcache

import (
	"context"
	"reflect"
	"time"

	"github.com/wencan/fastrest/resterror"
	"github.com/wencan/gox/xsync/sentinel"
)

// Storage 缓存存储接口。一般是对接redis、lru，透明处理业务数据。
// 因为MStorage和Storage的数据可能存储在一起，当key格式相同时，MStorage数据元素类型，应该同Storage数据元素类型。
type Storage interface {
	// Get 查询存储的数据。
	// 实现逻辑应该处理掉需要忽略的错误。
	// 当key格式相同时，valuePtr指针参数指向的类型，同Set方法的value参数的类型。
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

	// Query 如果没从Storage里找到，调用Query查询。
	Query QueryFunc

	// TTLRange 缓存生存时间区间。每次随机取一个区间内的值。
	TTLRange [2]time.Duration

	// sentinelGroup  哨兵机制。
	sentinelGroup sentinel.SentinelGroup

	// SentinelTTL 哨兵和哨兵持有的临时缓存的生存时间。用来省去双重检查。一般1s即可。
	// “副作用”是可以避免高频查询找不到的数据。
	SentinelTTL time.Duration
}

// Get 查询。destPtr为结果对象指针，key为缓存key，args为查询函数参数。
// destPtr的值是共享的，内容数据不可修改。
func (caching *Caching) Get(ctx context.Context, destPtr interface{}, key string, args interface{}) (found bool, err error) {
	// 先查缓存
	found, err = caching.Storage.Get(ctx, key, destPtr)
	if err != nil {
		// 错误
		// Storage的实现逻辑应该处理掉需要忽略的错误
		return false, err
	} else if found {
		// 通过可选的Validatable接口检查是否失效
		// 如果已经失效， destPtr指向已污染的数据。如果实现了Reset方法，执行Reset
		valid := validateAndReset(destPtr)
		if valid {
			return true, nil
		}
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
		dest := reflect.ValueOf(destPtr).Elem().Interface() // 传入指针，是为了取得值。这里存指针指向的内容。
		err = caching.Storage.Set(ctx, key, dest, getTTL(caching.TTLRange))
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

// validateAndReset 校验缓存对象。如果失效且支持Reset，就Reset
func validateAndReset(destPtr interface{}) (valid bool) {
	validator, _ := destPtr.(Validatable)
	if validator == nil || validator.IsValidCache() {
		// 没实现Validatable接口，或者实现了但是检查结果是有效
		return true
	}
	// 如果不是有效的，等同没找到
	// 这里destPtr已经被污染，指向一个失效的缓存对象
	resetable, _ := destPtr.(Resetable)
	if resetable != nil {
		resetable.Reset()
	}

	return false
}
