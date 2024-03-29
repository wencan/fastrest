package lrucache

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLRUCache(t *testing.T) {
	type Response struct {
		Echo string
	}

	lruCache := NewLRUCache(100, 10)

	// not found
	var resp1 Response
	ok, err := lruCache.Get(context.TODO(), "response_1", &resp1)
	if assert.Nil(t, err) {
		assert.False(t, ok)
	}

	err = lruCache.Set(context.TODO(), "response_1", Response{Echo: "response_1"}, time.Millisecond*100)
	assert.Nil(t, err)

	// got
	var resp2 Response
	ok, err = lruCache.Get(context.TODO(), "response_1", &resp2)
	if assert.Nil(t, err) {
		if assert.True(t, ok) {
			assert.Equal(t, "response_1", resp2.Echo)
		}
	}

	// expired
	time.Sleep(time.Millisecond * 300)
	var resp3 Response
	ok, err = lruCache.Get(context.TODO(), "response_1", &resp3)
	if assert.Nil(t, err) {
		assert.False(t, ok)
	}

	keys := []string{"response_1", "response_2", "response_3"}
	values := []*Response{{Echo: "response_1"}, {Echo: "response_2"}, {Echo: "response_3"}}
	err = lruCache.MSet(context.TODO(), keys, values, time.Minute)
	assert.Nil(t, err)

	// found all
	keys = []string{"response_1", "response_2", "response_3"}
	var resp4 []*Response
	missIndexed, err := lruCache.MGet(context.TODO(), keys, &resp4)
	if assert.Nil(t, err) {
		assert.Empty(t, missIndexed)
		assert.Equal(t, []*Response{{Echo: "response_1"}, {Echo: "response_2"}, {Echo: "response_3"}}, resp4)
	}

	//  partial notfound
	keys = []string{"response_1", "response_2", "response_notfound", "response_3", "response_miss"}
	var resp5 []*Response
	missIndexed, err = lruCache.MGet(context.TODO(), keys, &resp5)
	if assert.Nil(t, err) {
		assert.Equal(t, []int{2, 4}, missIndexed)
		assert.Equal(t, []*Response{{Echo: "response_1"}, {Echo: "response_2"}, {Echo: "response_3"}}, resp5)
	}

	// all notfound
	keys = []string{"response_notfound", "response_miss"}
	var resp6 []*Response
	missIndexed, err = lruCache.MGet(context.TODO(), keys, &resp6)
	if assert.Nil(t, err) {
		assert.Equal(t, []int{0, 1}, missIndexed)
		assert.Empty(t, resp6)
	}
}

func TestLRUCache_Concurrently(t *testing.T) {
	lruCache := NewLRUCache(1000, 10)

	ctx := context.TODO()
	for i := 0; i < 10000; i++ {
		err := lruCache.Set(ctx, strconv.Itoa(i), i, time.Minute)
		assert.Nil(t, err)
	}

	var wg sync.WaitGroup
	for idx := 0; idx < 100; idx++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < 10000; j++ {
				number := rand.Intn(10000)
				err := lruCache.Set(ctx, strconv.Itoa(number), number, time.Minute)
				if !assert.Nil(t, err) {
					return
				}

				number = rand.Intn(10000)
				key := strconv.Itoa(number)
				var got int
				want := number
				found, err := lruCache.Get(ctx, key, &got)
				if !assert.Nil(t, err) {
					return
				}
				if found {
					// 可以没找到，但不能出错
					assert.Equal(t, want, got)
				}
			}
		}()
	}

	wg.Wait()
}
