package restcache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wencan/fastrest/restcache/lrucache"
)

func ExampleCaching_Get() {
	query := func(ctx context.Context, destPtr interface{}, args interface{}) (found bool, err error) {
		req := args.(string)
		resp := destPtr.(*string)
		*resp = "echo: " + req
		return true, nil
	}

	caching := Caching{
		Storage:     lrucache.NewLRUCache(10000, 10), // 一般是redis实现
		Query:       query,
		TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
		SentinelTTL: time.Second,
	}

	var resp string
	found, err := caching.Get(context.TODO(), &resp, "key:hello", "hello")
	if err != nil {
		fmt.Println(err)
		return
	}
	if found {
		fmt.Println(resp)
	}

	// Output: echo: hello
}

func ExampleMCaching_MGet() {
	query := func(ctx context.Context, destSlicePtr interface{}, argsSlice interface{}) (missIndexes []int, err error) {
		req := argsSlice.([]string)
		resp := destSlicePtr.(*[]string)
		for _, r := range req {
			*resp = append(*resp, "echo: "+r)
		}
		return nil, nil
	}

	mcaching := MCaching{
		MStorage:    lrucache.NewLRUCache(10000, 10), // 一般是redis实现
		MQuery:      query,
		TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
		SentinelTTL: time.Second,
	}

	var keys = []string{"key_1", "key_2", "key_3"}
	var args = []string{"1", "2", "3"}
	var resp []string
	_, err := mcaching.MGet(context.TODO(), &resp, keys, args)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(strings.Join(resp, "\n"))

	// Output: echo: 1
	// echo: 2
	// echo: 3
}

func ExampleMCaching_partialFound() {
	query := func(ctx context.Context, destSlicePtr interface{}, argsSlice interface{}) (missIndexes []int, err error) {
		req := argsSlice.([]string)
		resp := destSlicePtr.(*[]string)
		for index, r := range req {
			if r == "" {
				missIndexes = append(missIndexes, index)
				continue
			}
			*resp = append(*resp, "echo: "+r)
		}
		return missIndexes, nil
	}

	mcaching := MCaching{
		MStorage:    lrucache.NewLRUCache(10000, 10), // 一般是redis实现
		MQuery:      query,
		TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
		SentinelTTL: time.Second,
	}

	var keys = []string{"key_1", "key_2", "key_notfound", "key_3"}
	var args = []string{"1", "2", "", "3"}
	var resp []string
	missIndexes, err := mcaching.MGet(context.TODO(), &resp, keys, args)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(strings.Join(resp, "; "))
	for _, missIndex := range missIndexes {
		fmt.Println("not found:", keys[missIndex])
	}

	// Output: echo: 1; echo: 2; echo: 3
	// not found: key_notfound
}

func ExampleStorage_commonStorage() {
	commonStorage := lrucache.NewLRUCache(10000, 10)

	query := func(ctx context.Context, destPtr interface{}, args interface{}) (found bool, err error) {
		req := args.(string)
		resp := destPtr.(*string)
		*resp = "echo: " + req
		return true, nil
	}
	mquery := func(ctx context.Context, destSlicePtr interface{}, argsSlice interface{}) (missIndexes []int, err error) {
		req := argsSlice.([]string)
		resp := destSlicePtr.(*[]string)
		for _, r := range req {
			*resp = append(*resp, "echo: "+r)
		}
		return nil, nil
	}

	caching := Caching{
		Storage:     commonStorage,
		Query:       query,
		TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
		SentinelTTL: time.Second,
	}
	mcaching := MCaching{
		MStorage:    commonStorage,
		MQuery:      mquery,
		TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
		SentinelTTL: time.Second,
	}

	var resp string
	key := "key:abc"
	args := "abc"
	found, err := caching.Get(context.TODO(), &resp, key, args)
	if err != nil {
		fmt.Println(err)
		return
	}
	if found {
		fmt.Println(resp)
	}

	var keys = []string{"key:abc", "key:def"}
	var argsSlice = []string{"abc", "def"}
	var respSlice []string
	_, err = mcaching.MGet(context.TODO(), &respSlice, keys, argsSlice)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(strings.Join(respSlice, "\n"))

	// Output: echo: abc
	// echo: abc
	// echo: def
}
