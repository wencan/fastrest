package restcache

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/wencan/fastrest/resterror"
	"github.com/wencan/fastrest/restutils"
	"github.com/wencan/gox/xsync/sentinel"
)

// MStorage 支持批量操作的缓存存储接口。一般是对接redis、lru，透明处理业务数据。
type MStorage interface {
	// MGet 批量查询存储的数据。
	// destSlicePtr元素的顺序同keys的顺序。如果keys重复，destSlicePtr元素也必须相应重复。
	// 如果全部没找到，或者部分没找到，返回没找到部分的下标，不返回错误。
	// 实现逻辑应该处理掉需要忽略的错误。
	MGet(ctx context.Context, keys []string, destSlicePtr interface{}) (missIndexes []int, err error)

	// MSet 批量存储数据。
	// keys和destSlice同长度同顺序。
	// 实现逻辑应该处理掉需要忽略的错误。
	MSet(ctx context.Context, keys []string, destSlice interface{}, ttl time.Duration) error
}

// MQueryFunc 批量查询函数。
// destSlicePtr元素的顺序同argsSlice的顺序。如果argsSlice元素出现重复，destSlicePtr元素也必须相应重复。
// 如果全部没找到，或者部分没找到，返回没找到部分的下标，不返回错误。
type MQueryFunc func(ctx context.Context, destSlicePtr interface{}, argsSlice interface{}) (missIndexes []int, err error)

type MCaching struct {
	// MStorage 支持批量操作的缓存存储接口。
	MStorage MStorage

	// MQuery 没从MStorage里找到的，调用MQuery批量查询。
	MQuery MQueryFunc

	// TTLRange 缓存生存时间区间。每次随机取一个区间内的值。
	TTLRange [2]time.Duration

	// sentinelGroup  哨兵机制。
	sentinelGroup sentinel.SentinelGroup

	// SentinelTTL 哨兵和哨兵持有的临时缓存的生存时间。用来省去双重检查。一般1s即可。
	// “副作用”是可以避免高频查询找不到的数据。
	SentinelTTL time.Duration
}

// MGet 批量查询。
// destSlicePtr是结果切片指针，keys是缓存key，argsSlice是查询函数的参数序列。
// 如果全部没找到，或者部分没找到，返回没找到部分的下标，不返回错误。
// destSlicePtr指向的切片的元素数据是共享的，内容不可修改。
func (mcaching *MCaching) MGet(ctx context.Context, destSlicePtr interface{}, keys []string, argsSlice interface{}) (missIndexes []int, err error) {
	// 第一步，先查缓存
	cacheMissIndexes, err := mcaching.MStorage.MGet(ctx, keys, destSlicePtr)
	if err != nil {
		return nil, err
	}
	if len(cacheMissIndexes) == 0 {
		// 全部找到
		return nil, nil
	}

	// 第二步，调query查询函数，查缓存没命中的
	// 未命中缓存的查询参数
	missKeys := make([]string, 0, len(cacheMissIndexes))
	argsSliceValue := reflect.ValueOf(argsSlice)
	if len(keys) != argsSliceValue.Len() {
		return nil, errors.New("wrong argsSlice")
	}
	missArgsSliceValue := reflect.MakeSlice(argsSliceValue.Type(), 0, len(cacheMissIndexes))
	for _, missIndex := range cacheMissIndexes {
		missKeys = append(missKeys, keys[missIndex])
		missArgsSliceValue = reflect.Append(missArgsSliceValue, argsSliceValue.Index(missIndex))
	}
	// query查询
	queriedDestPtrValue := reflect.New(reflect.ValueOf(destSlicePtr).Type().Elem())
	// 哨兵机制。同一进程内，同一时间，不同查询同key的数据
	queryErrs, err := mcaching.sentinelGroup.MDo(ctx, queriedDestPtrValue.Interface(), missKeys, missArgsSliceValue.Interface(), func(ctx context.Context, destSlicePtr, argsSlice interface{}) ([]error, error) {
		queryMissIndexes, err := mcaching.MQuery(ctx, destSlicePtr, argsSlice)
		if err != nil {
			return nil, err
		}
		var errs []error // 目前这个只会有notfound或者nil
		for index, missKey := range missKeys {
			if restutils.IntSliceContains(queryMissIndexes, index) {
				errs = append(errs, resterror.FormatNotFoundError("not found key [%s]", missKey))
			} else {
				errs = append(errs, nil)
			}
		}
		return errs, nil
	})
	if err != nil {
		return nil, err
	}
	// 延迟删除哨兵（和哨兵持有的临时缓存）
	// 省去双重检查。
	time.AfterFunc(mcaching.SentinelTTL, func() {
		mcaching.sentinelGroup.Delete(missKeys...)
	})

	// 第三步，query查询到的存起来
	queriedDestValue := queriedDestPtrValue.Elem()
	var queriedKeys = make([]string, 0, queriedDestValue.Len())
	for queryIndex, missKey := range missKeys {
		if len(queryErrs) > queryIndex && resterror.IsNotFound(queryErrs[queryIndex]) {
			// query函数没找到
		} else {
			queriedKeys = append(queriedKeys, missKey)
		}
	}
	if len(queriedKeys) != queriedDestValue.Len() {
		return nil, fmt.Errorf("wrong query result. query keys: %v", missKeys)
	}
	err = mcaching.MStorage.MSet(ctx, queriedKeys, queriedDestValue.Interface(), getTTL(mcaching.TTLRange))
	if err != nil {
		return nil, err
	}

	// 第四步，组合结果
	var cacheCount, queryCount, queriedDestCount int
	var cacheDestSlice = reflect.ValueOf(destSlicePtr).Elem()
	var destElemValueMap = make(map[string]reflect.Value)
	var m = make(map[string]interface{})
	for index, key := range keys {
		if !restutils.IntSliceContains(cacheMissIndexes, index) { // 缓存命中的
			if cacheDestSlice.Len() <= cacheCount { // 缓存返回数据有问题
				return nil, errors.New("not enough cache results")
			}
			// fmt.Println(key, cacheDestSlice.Index(cacheCount).Interface())
			destElemValueMap[key] = cacheDestSlice.Index(cacheCount)
			m[key] = cacheDestSlice.Index(cacheCount).Interface()
			cacheCount++
		} else { // 缓存没命中的
			if len(queryErrs) > queryCount && resterror.IsNotFound(queryErrs[queryCount]) { // 允许省去后面的nil。目前queryErrs只会有notfound或者nil
				// query函数也没找到
				missIndexes = append(missIndexes, index)
			} else { // 查询到的
				if queriedDestValue.Len() <= queriedDestCount { // query函数返回数据有问题
					return nil, errors.New("not enough query results")
				}
				// fmt.Println(key, queriedDestValue.Index(queriedDestCount).Interface())
				destElemValueMap[key] = queriedDestValue.Index(queriedDestCount)
				m[key] = queriedDestValue.Index(queriedDestCount).Interface()
				queriedDestCount++
			}
			queryCount++
		}
	}
	// 给destSlicePtr赋值
	destSliceValue := reflect.ValueOf(destSlicePtr).Elem()
	destSliceValue.Set(reflect.MakeSlice(destSliceValue.Type(), 0, len(keys))) // 之前缓存append过数据
	for _, key := range keys {
		destValue, ok := destElemValueMap[key]
		if !ok { // not found的项
			continue
		}
		destSliceValue.Set(reflect.Append(destSliceValue, destValue))
	}

	return missIndexes, nil
}
