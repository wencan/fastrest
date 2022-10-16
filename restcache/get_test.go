package restcache

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/wencan/fastrest/restcache/mock_restcache"
)

func TestCaching_GetCached(t *testing.T) {
	// 直接从缓存查到

	type Response struct {
		Echo string
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	mockStorage := mock_restcache.NewMockStorage(ctrl)
	mockStorage.EXPECT().Get(gomock.Eq(ctx), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf((*Response)(nil))).DoAndReturn(func(ctx context.Context, key string, destPtr interface{}) (found bool, err error) {
		p := destPtr.(*Response)
		p.Echo = "echo: " + key
		return true, nil
	}).Times(1)

	caching := Caching{
		Storage:  mockStorage,
		Query:    nil,
		TTLRange: [2]time.Duration{time.Minute * 4, time.Minute * 6},
	}
	var response Response
	found, err := caching.Get(ctx, &response, "hi", nil)
	if assert.Nil(t, err) {
		if assert.True(t, found) {
			assert.Equal(t, "echo: hi", response.Echo)
		}
	}
}

func TestCaching_GetQueried(t *testing.T) {
	// 未命中缓存，从query查询
	// 存到缓存。
	// 然后从直接从缓存查到

	type Response struct {
		Echo string
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	mockInternalStore := make(map[string]interface{})
	mockStorage := mock_restcache.NewMockStorage(ctrl)
	mockStorage.EXPECT().Get(gomock.Eq(ctx), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf((*Response)(nil))).DoAndReturn(func(ctx context.Context, key string, destPtr interface{}) (found bool, err error) {
		value, ok := mockInternalStore[key]
		if !ok {
			return false, nil
		}
		p := destPtr.(*Response)
		*p = value.(Response)
		return true, nil
	}).Times(2)
	mockStorage.EXPECT().Set(gomock.Eq(ctx), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf((*Response)(nil)), gomock.AssignableToTypeOf(time.Second)).DoAndReturn(func(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
		response := value.(*Response)
		mockInternalStore[key] = *response
		return nil
	}).Times(1)

	query := func(ctx context.Context, destPtr interface{}, args interface{}) (found bool, err error) {
		s := args.(string)
		response := destPtr.(*Response)
		response.Echo = "echo: " + s
		return true, nil
	}

	caching := Caching{
		Storage:  mockStorage,
		Query:    query,
		TTLRange: [2]time.Duration{time.Minute * 4, time.Minute * 6},
	}
	var response Response
	found, err := caching.Get(ctx, &response, "hi", "hi")
	if assert.Nil(t, err) {
		if assert.True(t, found) {
			assert.Equal(t, "echo: hi", response.Echo)
		}
	}

	// 直接命中缓存
	var response2 Response
	found, err = caching.Get(ctx, &response2, "hi", "hi")
	if assert.Nil(t, err) {
		if assert.True(t, found) {
			assert.Equal(t, "echo: hi", response.Echo)
		}
	}
}

func TestCaching_ConcurrentlyGet(t *testing.T) {
	type Response struct {
		Echo string
	}

	var big = 1000

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	mockInternalStore := sync.Map{}
	mockStorage := mock_restcache.NewMockStorage(ctrl)
	mockStorage.EXPECT().Get(gomock.Eq(ctx), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf((*Response)(nil))).DoAndReturn(func(ctx context.Context, key string, destPtr interface{}) (found bool, err error) {
		value, ok := mockInternalStore.Load(key)
		if !ok {
			return false, nil
		}
		p := destPtr.(*Response)
		p.Echo = value.(string)
		return true, nil
	}).AnyTimes()
	mockStorage.EXPECT().Set(gomock.Eq(ctx), gomock.AssignableToTypeOf(""), gomock.AssignableToTypeOf((*Response)(nil)), gomock.AssignableToTypeOf(time.Second)).DoAndReturn(func(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
		response := value.(*Response)
		mockInternalStore.Store(key, response.Echo)
		return nil
	}).Times(big)

	var counts = make([]uint64, big)
	query := func(ctx context.Context, dest interface{}, args interface{}) (found bool, err error) {
		index := args.(int)

		// 每个索引只执行一次
		count := atomic.AddUint64(&counts[index], 1)
		if count != 1 {
			return false, fmt.Errorf("conflict, index: %d, count: %d", index, count)
		}

		response := dest.(*Response)
		response.Echo = "echo: " + strconv.Itoa(index)
		return true, nil
	}

	caching := Caching{
		Storage:     mockStorage,
		Query:       query,
		TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
		SentinelTTL: time.Second * 1,
	}

	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup
	wg.Add(500)
	for i := 0; i < 500; i++ {
		go func() {
			defer wg.Done()

			for index := range rand.Perm(big) {
				key := strconv.Itoa(index)

				var response Response
				found, err := caching.Get(ctx, &response, key, index)
				if !assert.Nil(t, err) {
					continue
				}
				if assert.True(t, found) {
					assert.Equal(t, "echo: "+key, response.Echo)
				}
			}

		}()
	}
	wg.Wait()
}

func TestCaching_Get(t *testing.T) {
	type Request struct {
		Greetings string
	}
	type Response struct {
		Echo string
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 能正常工作的Storage
	mockInternalStore := sync.Map{}
	mockStorage := mock_restcache.NewMockStorage(ctrl)
	mockStorage.EXPECT().Get(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf(""), gomock.Any()).DoAndReturn(func(ctx context.Context, key string, destPtr interface{}) (found bool, err error) {
		value, ok := mockInternalStore.Load(key)
		if !ok {
			return false, nil
		}
		reflect.ValueOf(destPtr).Elem().Set(reflect.ValueOf(value))
		return true, nil
	}).AnyTimes()
	mockStorage.EXPECT().Set(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf(""), gomock.Any(), gomock.AssignableToTypeOf(time.Second)).DoAndReturn(func(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
		mockInternalStore.Store(key, value)
		return nil
	}).AnyTimes()

	// 只会报not found的Storage
	notfoundStorage := mock_restcache.NewMockStorage(ctrl)
	notfoundStorage.EXPECT().Get(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf(""), gomock.Any()).DoAndReturn(func(ctx context.Context, key string, destPtr interface{}) (found bool, err error) {
		return false, nil
	}).AnyTimes()
	notfoundStorage.EXPECT().Set(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf(""), gomock.Any(), gomock.AssignableToTypeOf(time.Second)).DoAndReturn(func(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
		return nil
	}).AnyTimes()

	// 只会报错的Storage
	wrongStorage := mock_restcache.NewMockStorage(ctrl)
	wrongStorage.EXPECT().Get(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf(""), gomock.Any()).DoAndReturn(func(ctx context.Context, key string, destPtr interface{}) (found bool, err error) {
		return false, errors.New("Fail in get")
	}).AnyTimes()
	wrongStorage.EXPECT().Set(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf(""), gomock.Any(), gomock.AssignableToTypeOf(time.Second)).DoAndReturn(func(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
		return errors.New("Fail in set")
	}).AnyTimes()

	type fields struct {
		Storage     Storage
		Query       QueryFunc
		TTLRange    [2]time.Duration
		SentinelTTL time.Duration
	}
	type args struct {
		newDestPtr func() interface{}
		key        string
		args       interface{}
	}
	tests := []struct {
		name       string
		goroutines int
		fields     fields
		args       args
		want       interface{}
		wantFound  bool
		wantErr    bool
	}{
		{
			name:       "query_string", // 没命中缓存，查询到
			goroutines: 1000,
			fields: fields{
				Storage: notfoundStorage,
				Query: func(ctx context.Context, destPtr, args interface{}) (found bool, err error) {
					req := args.(string)
					resp := destPtr.(*string)
					*resp = "echo: " + req
					return true, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second,
			},
			args: args{
				newDestPtr: func() interface{} {
					return new(string)
				},
				key:  "no_hit_string",
				args: "no_hit",
			},
			want:      "echo: no_hit",
			wantFound: true,
		},
		{
			name:       "query_cached_string", // 命中缓存
			goroutines: 1000,
			fields: fields{
				Storage: func() Storage {
					mockInternalStore.Store("hit_cache_string", "echo: hit_cache") // 先存缓存
					return mockStorage
				}(),
				Query:       nil,
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second,
			},
			args: args{
				newDestPtr: func() interface{} {
					return new(string)
				},
				key:  "hit_cache_string",
				args: "hit_cache",
			},
			want:      "echo: hit_cache",
			wantFound: true,
		},
		{
			name:       "query_struct", // 没命中缓存，查询到
			goroutines: 1000,
			fields: fields{
				Storage: notfoundStorage,
				Query: func(ctx context.Context, destPtr, args interface{}) (found bool, err error) {
					req := args.(Request)
					resp := destPtr.(*Response)
					resp.Echo = "echo: " + req.Greetings
					return true, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second,
			},
			args: args{
				newDestPtr: func() interface{} {
					return new(Response)
				},
				key:  "no_hit_struct",
				args: Request{Greetings: "no_hit"},
			},
			want:      Response{Echo: "echo: no_hit"},
			wantFound: true,
		},
		{
			name:       "query_cached_struct", // 命中缓存
			goroutines: 1000,
			fields: fields{
				Storage: func() Storage {
					mockInternalStore.Store("hit_cache_struct", Response{Echo: "echo: hit_cache"}) // 先存缓存
					return mockStorage
				}(),
				Query:       nil,
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second,
			},
			args: args{
				newDestPtr: func() interface{} {
					return new(Response)
				},
				key:  "hit_cache_struct",
				args: Request{Greetings: "hit_cache"},
			},
			want:      Response{Echo: "echo: hit_cache"},
			wantFound: true,
		},
		{
			name:       "query_struct_ptr", // 没命中缓存，查询到
			goroutines: 1000,
			fields: fields{
				Storage: notfoundStorage,
				Query: func(ctx context.Context, destPtr, args interface{}) (found bool, err error) {
					req := args.(*Request)
					resp := destPtr.(**Response)
					*resp = &Response{Echo: "echo: " + req.Greetings}
					return true, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second,
			},
			args: args{
				newDestPtr: func() interface{} {
					return new(*Response)
				},
				key:  "no_hit_struct_ptr",
				args: &Request{Greetings: "no_hit_ptr"},
			},
			want:      &Response{Echo: "echo: no_hit_ptr"},
			wantFound: true,
		},
		{
			name:       "query_cached_struct_ptr", // 命中缓存
			goroutines: 1000,
			fields: fields{
				Storage: func() Storage {
					mockInternalStore.Store("hit_cache_struct_ptr", &Response{Echo: "echo: hit_cache_ptr"}) // 先存缓存
					return mockStorage
				}(),
				Query:       nil,
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second,
			},
			args: args{
				newDestPtr: func() interface{} {
					return new(*Response)
				},
				key:  "hit_cache_struct_ptr",
				args: Request{Greetings: "hit_cache_ptr"},
			},
			want:      &Response{Echo: "echo: hit_cache_ptr"},
			wantFound: true,
		},
		{
			name:       "query_notfound", // 查不到
			goroutines: 1000,
			fields: fields{
				Storage: mockStorage,
				Query: func(ctx context.Context, destPtr, args interface{}) (found bool, err error) {
					return false, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second,
			},
			args: args{
				newDestPtr: func() interface{} {
					return new(Response)
				},
				key:  "query_notfound",
				args: Request{Greetings: "notfound"},
			},
			wantFound: false,
		},
		{
			name:       "query_error", // 查询错误
			goroutines: 1000,
			fields: fields{
				Storage: mockStorage,
				Query: func(ctx context.Context, destPtr, args interface{}) (found bool, err error) {
					return false, errors.New("wow")
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second,
			},
			args: args{
				newDestPtr: func() interface{} {
					return new(Response)
				},
				key:  "query_error",
				args: Request{Greetings: "query_error"},
			},
			wantErr: true,
		},
		{
			name:       "cache_error", // 缓存错误
			goroutines: 1000,
			fields: fields{
				Storage: wrongStorage,
				Query: func(ctx context.Context, destPtr, args interface{}) (found bool, err error) {
					req := args.(Request)
					resp := destPtr.(*Response)
					resp.Echo = "echo: " + req.Greetings
					return true, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second,
			},
			args: args{
				newDestPtr: func() interface{} {
					return new(Response)
				},
				key:  "cache_error",
				args: Request{Greetings: "cache_error"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caching := &Caching{
				Storage:     tt.fields.Storage,
				Query:       tt.fields.Query,
				TTLRange:    tt.fields.TTLRange,
				SentinelTTL: tt.fields.SentinelTTL,
			}

			var wg sync.WaitGroup
			wg.Add(tt.goroutines)
			for i := 0; i < tt.goroutines; i++ {
				go func() {
					defer wg.Done()

					destPtr := tt.args.newDestPtr()
					gotFound, err := caching.Get(context.TODO(), destPtr, tt.args.key, tt.args.args)
					if tt.wantErr {
						assert.NotNil(t, err)
						return
					} else {
						if assert.Nil(t, err) {
							if tt.wantFound {
								assert.Equal(t, tt.want, reflect.ValueOf(destPtr).Elem().Interface())
							} else {
								assert.False(t, gotFound)
							}
						}
					}
				}()
			}
			wg.Wait()
		})
	}
}
