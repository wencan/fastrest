package restcache

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/wencan/fastrest/restcache/mock_restcache"
)

func TestMCaching_SimpleMGet(t *testing.T) {
	type Request struct {
		Greetings string
	}
	type Response struct {
		Echo string
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 存储
	mockStorage := mock_restcache.NewMockMStorage(ctrl)
	mockInternalStore := sync.Map{}
	mockStorage.EXPECT().MGet(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf([]string{}), gomock.Any()).DoAndReturn(func(ctx context.Context, keys []string, destSlicePtr interface{}) (missIndexes []int, err error) {
		var destSliceValue = reflect.ValueOf(destSlicePtr).Elem()
		for index, key := range keys {
			v, ok := mockInternalStore.Load(key)
			if ok {
				destSliceValue.Set(reflect.Append(destSliceValue, reflect.ValueOf(v)))
			} else {
				missIndexes = append(missIndexes, index)
			}
		}
		return missIndexes, nil
	}).AnyTimes()
	mockStorage.EXPECT().MSet(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf([]string{}), gomock.Any(), gomock.AssignableToTypeOf(time.Second)).DoAndReturn(func(ctx context.Context, keys []string, destSlice interface{}, ttl time.Duration) error {
		var destSliceValue = reflect.ValueOf(destSlice)
		for index, key := range keys {
			mockInternalStore.Store(key, destSliceValue.Index(index).Interface())
		}
		return nil
	}).AnyTimes()

	// 查询函数
	// 如果Greetings参数为空，算notfound
	mquery := func(ctx context.Context, destSlicePtr, argsSlice interface{}) (missIndexes []int, err error) {
		reqs := argsSlice.([]*Request)
		resps := destSlicePtr.(*[]*Response)

		for index, req := range reqs {
			if req.Greetings == "" { // 如果Greetings参数为空，算notfound
				missIndexes = append(missIndexes, index)
				continue
			}
			*resps = append(*resps, &Response{
				Echo: "echo: " + req.Greetings,
			})
		}
		return missIndexes, nil
	}

	mcaching := MCaching{
		MStorage:    mockStorage,
		MQuery:      mquery,
		TTLRange:    [2]time.Duration{time.Minute * 5, time.Minute * 6},
		SentinelTTL: time.Second * 5,
	}

	// 一个没缓存，但能查到的
	var resps1 []*Response
	_, err := mcaching.MGet(context.TODO(), &resps1, []string{"cached1"}, []*Request{{Greetings: "1"}})
	if assert.Nil(t, err) {
		assert.Equal(t, []*Response{{Echo: "echo: 1"}}, resps1)
	}

	// 查到上面缓存的
	var resps2 []*Response
	_, err = mcaching.MGet(context.TODO(), &resps2, []string{"cached1"}, []*Request{{Greetings: "1"}})
	if assert.Nil(t, err) {
		assert.Equal(t, []*Response{{Echo: "echo: 1"}}, resps2)
	}

	// 查到上面缓存的，和一个query到的
	var resps3 []*Response
	_, err = mcaching.MGet(context.TODO(), &resps3, []string{"cached1", "queried2"}, []*Request{{Greetings: "1"}, {Greetings: "2"}})
	if assert.Nil(t, err) {
		assert.Equal(t, []*Response{{Echo: "echo: 1"}, {Echo: "echo: 2"}}, resps3)
	}

	// 查到一个query到的
	var resps4 []*Response
	_, err = mcaching.MGet(context.TODO(), &resps4, []string{"queried3"}, []*Request{{Greetings: "3"}})
	if assert.Nil(t, err) {
		assert.Equal(t, []*Response{{Echo: "echo: 3"}}, resps4)
	}

	// 查到上面缓存的，和一个notfound的
	var resps5 []*Response
	_, err = mcaching.MGet(context.TODO(), &resps5, []string{"cached1", "queried2", "queried3", "notfound"}, []*Request{{Greetings: "1"}, {Greetings: "2"}, {Greetings: "3"}, {Greetings: ""}})
	if assert.Nil(t, err) {
		assert.Equal(t, []*Response{{Echo: "echo: 1"}, {Echo: "echo: 2"}, {Echo: "echo: 3"}}, resps5)
	}
}

func TestCaching_ConcurrentlyMGet(t *testing.T) {
	type Request struct {
		Index int
	}
	type Response struct {
		Echo string
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock_restcache.NewMockMStorage(ctrl)
	mockInternalStore := sync.Map{}
	mockStorage.EXPECT().MGet(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf([]string{}), gomock.Any()).DoAndReturn(func(ctx context.Context, keys []string, destSlicePtr interface{}) (missIndexes []int, err error) {
		var destSliceValue = reflect.ValueOf(destSlicePtr).Elem()
		for index, key := range keys {
			v, ok := mockInternalStore.Load(key)
			if ok {
				destSliceValue.Set(reflect.Append(destSliceValue, reflect.ValueOf(v)))
			} else {
				missIndexes = append(missIndexes, index)
			}
		}
		return missIndexes, nil
	}).AnyTimes()
	mockStorage.EXPECT().MSet(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf([]string{}), gomock.Any(), gomock.AssignableToTypeOf(time.Second)).DoAndReturn(func(ctx context.Context, keys []string, destSlice interface{}, ttl time.Duration) error {
		var destSliceValue = reflect.ValueOf(destSlice)
		for index, key := range keys {
			mockInternalStore.Store(key, destSliceValue.Index(index).Interface())
		}
		return nil
	}).AnyTimes()

	big := 10000
	counts := make([]uint64, big)
	mquery := func(ctx context.Context, destSlicePtr, argsSlice interface{}) (missIndexes []int, err error) {
		reqs := argsSlice.([]*Request)
		resps := destSlicePtr.(*[]*Response)

		for _, req := range reqs {
			count := atomic.AddUint64(&counts[req.Index], 1)
			*resps = append(*resps, &Response{
				Echo: fmt.Sprintf("index: %d, count: %d", req.Index, count),
			})
		}
		return missIndexes, nil
	}

	mcaching := MCaching{
		MStorage:    mockStorage,
		MQuery:      mquery,
		TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
		SentinelTTL: time.Second * 5,
	}

	var wg sync.WaitGroup
	wg.Add(500)
	for i := 0; i < 500; i++ {
		go func() {
			defer wg.Done()

			all := rand.Perm(big)
			var count int
			for {
				var keys []string
				var args []*Request
				var want []*Response

				for i := 0; i < rand.Intn(10); i++ {
					if count+1 > len(all) {
						break
					}
					index := all[count]
					name := fmt.Sprintf("index_%d", index)
					keys = append(keys, name)
					args = append(args, &Request{Index: index})
					want = append(want, &Response{Echo: fmt.Sprintf("index: %d, count: %d", index, 1)})

					count++
				}
				if big >= 0 && len(keys) == 0 {
					break // end
				}

				var resp []*Response
				_, err := mcaching.MGet(context.TODO(), &resp, keys, args)
				if assert.Nil(t, err) {
					if !assert.Equal(t, want, resp) {
						t.Logf("%+v, %+v", want, resp)
						return
					}
				}
			}
		}()
	}
	wg.Wait()
}

func TestMCaching_MGet(t *testing.T) {
	type Request struct {
		Greetings string
	}
	type Response struct {
		Echo string
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 能正常工作的Storage
	mockStorage := mock_restcache.NewMockMStorage(ctrl)
	mockInternalStore := sync.Map{}
	mockStorage.EXPECT().MGet(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf([]string{}), gomock.Any()).DoAndReturn(func(ctx context.Context, keys []string, destSlicePtr interface{}) (missIndexes []int, err error) {
		var destSliceValue = reflect.ValueOf(destSlicePtr).Elem()
		for index, key := range keys {
			v, ok := mockInternalStore.Load(key)
			if ok {
				destSliceValue.Set(reflect.Append(destSliceValue, reflect.ValueOf(v)))
			} else {
				missIndexes = append(missIndexes, index)
			}
		}
		return missIndexes, nil
	}).AnyTimes()
	mockStorage.EXPECT().MSet(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf([]string{}), gomock.Any(), gomock.AssignableToTypeOf(time.Second)).DoAndReturn(func(ctx context.Context, keys []string, destSlice interface{}, ttl time.Duration) error {
		var destSliceValue = reflect.ValueOf(destSlice)
		for index, key := range keys {
			mockInternalStore.Store(key, destSliceValue.Index(index).Interface())
		}
		return nil
	}).AnyTimes()

	// 只会返回notfound的Storage
	notfoundStorage := mock_restcache.NewMockMStorage(ctrl)
	notfoundStorage.EXPECT().MGet(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf([]string{}), gomock.Any()).DoAndReturn(func(ctx context.Context, keys []string, destSlicePtr interface{}) (missIndexes []int, err error) {
		for index := range keys {
			missIndexes = append(missIndexes, index)
		}
		return missIndexes, nil
	}).AnyTimes()
	notfoundStorage.EXPECT().MSet(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf([]string{}), gomock.Any(), gomock.AssignableToTypeOf(time.Second)).DoAndReturn(func(ctx context.Context, keys []string, destSlice interface{}, ttl time.Duration) error {
		return nil
	}).AnyTimes()

	// 必定报错的Storage
	WrongStorage := mock_restcache.NewMockMStorage(ctrl)
	WrongStorage.EXPECT().MGet(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf([]string{}), gomock.Any()).DoAndReturn(func(ctx context.Context, keys []string, destSlicePtr interface{}) (missIndexes []int, err error) {
		return nil, errors.New("Fail in MGet")
	}).AnyTimes()
	WrongStorage.EXPECT().MSet(gomock.AssignableToTypeOf(context.TODO()), gomock.AssignableToTypeOf([]string{}), gomock.Any(), gomock.AssignableToTypeOf(time.Second)).DoAndReturn(func(ctx context.Context, keys []string, destSlice interface{}, ttl time.Duration) error {
		return errors.New("Fail in MSet")
	}).AnyTimes()

	type fields struct {
		MStorage    MStorage
		MQuery      MQueryFunc
		TTLRange    [2]time.Duration
		SentinelTTL time.Duration
	}
	type args struct {
		newDestSlicePtr func() interface{}
		keys            []string
		argsSlice       interface{}
	}
	tests := []struct {
		name            string
		goroutines      int
		fields          fields
		args            args
		want            interface{}
		wantMissIndexes []int
		wantErr         bool
	}{
		{
			name:       "one_cached_string",
			goroutines: 10000,
			fields: fields{
				MStorage: func() MStorage {
					mockInternalStore.Store("cached_string_1", "echo: 1")
					return mockStorage
				}(),
				MQuery:      nil,
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]string)
				},
				keys:      []string{"cached_string_1"},
				argsSlice: []int{1},
			},
			want: []string{"echo: 1"},
		},
		{
			name:       "multi_cached_string",
			goroutines: 10000,
			fields: fields{
				MStorage: func() MStorage {
					mockInternalStore.Store("cached_string_1", "echo: 1")
					mockInternalStore.Store("cached_string_2", "echo: 2")
					mockInternalStore.Store("cached_string_3", "echo: 3")
					mockInternalStore.Store("cached_string_4", "echo: 4")
					mockInternalStore.Store("cached_string_5", "echo: 5")
					mockInternalStore.Store("cached_string_6", "echo: 6")
					return mockStorage
				}(),
				MQuery:      nil,
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]string)
				},
				keys:      []string{"cached_string_1", "cached_string_2", "cached_string_3", "cached_string_4", "cached_string_5", "cached_string_6"},
				argsSlice: []int{1, 2, 3, 4, 5, 6},
			},
			want: []string{"echo: 1", "echo: 2", "echo: 3", "echo: 4", "echo: 5", "echo: 6"},
		},
		{
			name:       "one_queried_string",
			goroutines: 10000,
			fields: fields{
				MStorage: notfoundStorage,
				MQuery: func(ctx context.Context, destSlicePtr, argsSlice interface{}) (missIndexes []int, err error) {
					req := argsSlice.([]int)
					resp := destSlicePtr.(*[]string)
					for _, r := range req {
						*resp = append(*resp, fmt.Sprintf("return: %d", r))
					}
					return nil, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]string)
				},
				keys:      []string{"quired_string_1"},
				argsSlice: []int{1},
			},
			want: []string{"return: 1"},
		},
		{
			name:       "multi_queried_string",
			goroutines: 10000,
			fields: fields{
				MStorage: notfoundStorage,
				MQuery: func(ctx context.Context, destSlicePtr, argsSlice interface{}) (missIndexes []int, err error) {
					req := argsSlice.([]int)
					resp := destSlicePtr.(*[]string)
					for _, r := range req {
						*resp = append(*resp, fmt.Sprintf("return: %d", r))
					}
					return nil, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]string)
				},
				keys:      []string{"quired_string_2", "quired_string_3", "quired_string_4", "quired_string_5", "quired_string_6", "quired_string_7"},
				argsSlice: []int{2, 3, 4, 5, 6, 7},
			},
			want: []string{"return: 2", "return: 3", "return: 4", "return: 5", "return: 6", "return: 7"},
		},
		{
			name:       "multi_cached_struct",
			goroutines: 10000,
			fields: fields{
				MStorage: func() MStorage {
					mockInternalStore.Store("cached_struct_1", Response{Echo: "echo: 1"})
					mockInternalStore.Store("cached_struct_2", Response{Echo: "echo: 2"})
					mockInternalStore.Store("cached_struct_3", Response{Echo: "echo: 3"})
					return mockStorage
				}(),
				MQuery:      nil,
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]Response)
				},
				keys:      []string{"cached_struct_1", "cached_struct_2", "cached_struct_3"},
				argsSlice: []Request{{Greetings: "1"}, {Greetings: "2"}, {Greetings: "3"}},
			},
			want: []Response{{Echo: "echo: 1"}, {Echo: "echo: 2"}, {Echo: "echo: 3"}},
		},
		{
			name:       "multi_queried_struct",
			goroutines: 10000,
			fields: fields{
				MStorage: mockStorage,
				MQuery: func(ctx context.Context, destSlicePtr, argsSlice interface{}) (missIndexes []int, err error) {
					req := argsSlice.([]Request)
					resp := destSlicePtr.(*[]Response)
					for _, r := range req {
						*resp = append(*resp, Response{Echo: "echo: " + r.Greetings})
					}
					return nil, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]Response)
				},
				keys:      []string{"quired_struct_1", "quried_struct_2", "quired_struct_3"},
				argsSlice: []Request{{Greetings: "1"}, {Greetings: "2"}, {Greetings: "3"}},
			},
			want: []Response{{Echo: "echo: 1"}, {Echo: "echo: 2"}, {Echo: "echo: 3"}},
		},
		{
			name:       "multi_cached_struct_ptr",
			goroutines: 10000,
			fields: fields{
				MStorage: func() MStorage {
					mockInternalStore.Store("cached_struct_ptr_1", &Response{Echo: "echo: 1"})
					mockInternalStore.Store("cached_struct_ptr_2", &Response{Echo: "echo: 2"})
					mockInternalStore.Store("cached_struct_ptr_3", &Response{Echo: "echo: 3"})
					return mockStorage
				}(),
				MQuery:      nil,
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]*Response)
				},
				keys:      []string{"cached_struct_ptr_1", "cached_struct_ptr_2", "cached_struct_ptr_3"},
				argsSlice: []*Request{{Greetings: "1"}, {Greetings: "2"}, {Greetings: "3"}},
			},
			want: []*Response{{Echo: "echo: 1"}, {Echo: "echo: 2"}, {Echo: "echo: 3"}},
		},
		{
			name:       "multi_queried_struct_ptr",
			goroutines: 10000,
			fields: fields{
				MStorage: notfoundStorage,
				MQuery: func(ctx context.Context, destSlicePtr, argsSlice interface{}) (missIndexes []int, err error) {
					req := argsSlice.([]*Request)
					resp := destSlicePtr.(*[]*Response)
					for _, r := range req {
						*resp = append(*resp, &Response{Echo: "echo: " + r.Greetings})
					}
					return nil, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]*Response)
				},
				keys:      []string{"quired_struct_ptr_1", "quried_struct_ptr_2", "quired_struct_ptr_3"},
				argsSlice: []*Request{{Greetings: "1"}, {Greetings: "2"}, {Greetings: "3"}},
			},
			want: []*Response{{Echo: "echo: 1"}, {Echo: "echo: 2"}, {Echo: "echo: 3"}},
		},
		{
			name:       "one_cached_string-and-one_queried_string",
			goroutines: 10000,
			fields: fields{
				MStorage: func() MStorage {
					mockInternalStore.Store("cached_string_10", "echo: 10")
					return mockStorage
				}(),
				MQuery: func(ctx context.Context, destSlicePtr, argsSlice interface{}) (missIndexes []int, err error) {
					req := argsSlice.([]int)
					resp := destSlicePtr.(*[]string)
					for _, r := range req {
						*resp = append(*resp, fmt.Sprintf("return: %d", r))
					}
					return nil, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]string)
				},
				keys:      []string{"cached_string_10", "quired_string11"},
				argsSlice: []int{10, 11},
			},
			want: []string{"echo: 10", "return: 11"},
		},
		{
			name:       "multi_cached_string-and-multi_queried_string",
			goroutines: 10000,
			fields: fields{
				MStorage: func() MStorage {
					mockInternalStore.Store("cached_string_10", "echo: 10")
					mockInternalStore.Store("cached_string_11", "echo: 11")
					mockInternalStore.Store("cached_string_12", "echo: 12")
					return mockStorage
				}(),
				MQuery: func(ctx context.Context, destSlicePtr, argsSlice interface{}) (missIndexes []int, err error) {
					req := argsSlice.([]int)
					resp := destSlicePtr.(*[]string)
					for _, r := range req {
						*resp = append(*resp, fmt.Sprintf("return: %d", r))
					}
					return nil, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]string)
				},
				keys:      []string{"cached_string_10", "quired_string13", "cached_string_11", "cached_string_12", "quired_string14", "quired_string15"},
				argsSlice: []int{10, 13, 11, 12, 14, 15},
			},
			want: []string{"echo: 10", "return: 13", "echo: 11", "echo: 12", "return: 14", "return: 15"},
		},
		{
			name:       "multi_cached_string-and-multi_queried_string-2",
			goroutines: 10000,
			fields: fields{
				MStorage: func() MStorage {
					mockInternalStore.Store("cached_string_20", "echo: 20")
					mockInternalStore.Store("cached_string_21", "echo: 21")
					mockInternalStore.Store("cached_string_22", "echo: 22")
					return mockStorage
				}(),
				MQuery: func(ctx context.Context, destSlicePtr, argsSlice interface{}) (missIndexes []int, err error) {
					req := argsSlice.([]int)
					resp := destSlicePtr.(*[]string)
					for _, r := range req {
						*resp = append(*resp, fmt.Sprintf("return: %d", r))
					}
					return nil, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]string)
				},
				keys:      []string{"quired_string_25", "cached_string_20", "quired_string_26", "cached_string_21", "quired_string_27", "cached_string_22"},
				argsSlice: []int{25, 20, 26, 21, 27, 22},
			},
			want: []string{"return: 25", "echo: 20", "return: 26", "echo: 21", "return: 27", "echo: 22"},
		},
		{
			name:       "partial_notfound",
			goroutines: 10000,
			fields: fields{
				MStorage: mockStorage,
				MQuery: func(ctx context.Context, destSlicePtr, argsSlice interface{}) (missIndexes []int, err error) {
					req := argsSlice.([]int)
					resp := destSlicePtr.(*[]string)
					for index, r := range req {
						if r == 0 { // 入参为0的项，算not found
							missIndexes = append(missIndexes, index)
							continue
						}
						*resp = append(*resp, fmt.Sprintf("return: %d", r))
					}
					return missIndexes, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]string)
				},
				keys:      []string{"partial_notfound_1", "partial_notfound_2", "partial_notfound_0"},
				argsSlice: []int{1, 2, 0}, // 入参为0的项，会not found
			},
			want:            []string{"return: 1", "return: 2"},
			wantMissIndexes: []int{2},
		},
		{
			name:       "partial_notfound-2",
			goroutines: 10000,
			fields: fields{
				MStorage: mockStorage,
				MQuery: func(ctx context.Context, destSlicePtr, argsSlice interface{}) (missIndexes []int, err error) {
					req := argsSlice.([]int)
					resp := destSlicePtr.(*[]string)
					for index, r := range req {
						if r == 0 { // 入参为0的项，算not found
							missIndexes = append(missIndexes, index)
							continue
						}
						*resp = append(*resp, fmt.Sprintf("return: %d", r))
					}
					return missIndexes, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]string)
				},
				keys:      []string{"partial_notfound_1", "partial_notfound_01", "partial_notfound_02", "partial_notfound_2", "partial_notfound_03"},
				argsSlice: []int{1, 0, 0, 2, 0}, // 入参为0的项，会not found
			},
			want:            []string{"return: 1", "return: 2"},
			wantMissIndexes: []int{1, 2, 4},
		},
		{
			name:       "all_notfound",
			goroutines: 10000,
			fields: fields{
				MStorage: notfoundStorage,
				MQuery: func(ctx context.Context, destSlicePtr, argsSlice interface{}) (missIndexes []int, err error) {
					req := argsSlice.([]int)
					resp := destSlicePtr.(*[]string)
					for index, r := range req {
						if r == 0 { // 入参为0的项，算not found
							missIndexes = append(missIndexes, index)
							continue
						}
						*resp = append(*resp, fmt.Sprintf("return: %d", r))
					}
					return missIndexes, nil
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]string)
				},
				keys:      []string{"all_notfound_1", "all_notfound_2", "all_notfound_3", "all_notfound_4", "all_notfound_5"},
				argsSlice: []int{0, 0, 0, 0, 0}, // 入参为0的项，会not found
			},
			want:            []string{},
			wantMissIndexes: []int{0, 1, 2, 3, 4},
		},
		{
			name:       "cache_error",
			goroutines: 10000,
			fields: fields{
				MStorage:    WrongStorage,
				MQuery:      nil,
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]string)
				},
				keys:      []string{"cache_error"},
				argsSlice: []int{0},
			},
			want:    []string{},
			wantErr: true,
		},
		{
			name:       "query_error",
			goroutines: 10000,
			fields: fields{
				MStorage: mockStorage,
				MQuery: func(ctx context.Context, destSlicePtr, argsSlice interface{}) (missIndexes []int, err error) {
					return nil, errors.New("Fail in query")
				},
				TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
				SentinelTTL: time.Second * 5,
			},
			args: args{
				newDestSlicePtr: func() interface{} {
					return new([]string)
				},
				keys:      []string{"query_error"},
				argsSlice: []int{0},
			},
			want:    []string{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.goroutines == 0 {
				t.Fatal("no goroutines")
			}

			mcaching := &MCaching{
				MStorage:    tt.fields.MStorage,
				MQuery:      tt.fields.MQuery,
				TTLRange:    tt.fields.TTLRange,
				SentinelTTL: tt.fields.SentinelTTL,
			}

			var wg sync.WaitGroup
			wg.Add(tt.goroutines)
			for i := 0; i < tt.goroutines; i++ {
				go func() {
					defer wg.Done()

					destSlicePtr := tt.args.newDestSlicePtr()
					gotMissIndexes, err := mcaching.MGet(context.Background(), destSlicePtr, tt.args.keys, tt.args.argsSlice)
					if tt.wantErr {
						assert.NotNil(t, err)
					} else {
						assert.Equal(t, tt.wantMissIndexes, gotMissIndexes)
						assert.Equal(t, tt.want, reflect.ValueOf(destSlicePtr).Elem().Interface())
					}
				}()
			}
			wg.Wait()
		})
	}
}

func Test_removeInvalidCache(t *testing.T) {
	type args struct {
		keysLength       int
		cacheMissIndexes []int
		destSlicePtr     interface{}
	}
	tests := []struct {
		name                    string
		args                    args
		wantNewCacheMissIndexes []int
		wantErr                 bool
	}{
		{
			name: "not_support_validate",
			args: args{
				keysLength:       3,
				cacheMissIndexes: []int{0, 1},
				destSlicePtr:     &[]string{},
			},
			wantNewCacheMissIndexes: []int{0, 1},
		},
		{
			name: "all_found_valid",
			args: args{
				keysLength:       3,
				cacheMissIndexes: []int{},
				destSlicePtr:     &[]testResponseWithValid{{true}, {true}, {true}},
			},
			wantNewCacheMissIndexes: nil,
		},
		{
			name: "not_found_valid",
			args: args{
				keysLength:       3,
				cacheMissIndexes: []int{0, 1, 2},
				destSlicePtr:     &[]testResponseWithValid{},
			},
			wantNewCacheMissIndexes: []int{0, 1, 2},
		},
		{
			name: "all_found_invalid",
			args: args{
				keysLength:       3,
				cacheMissIndexes: []int{},
				destSlicePtr:     &[]testResponseWithValid{{false}, {false}, {false}},
			},
			wantNewCacheMissIndexes: []int{0, 1, 2},
		},
		{
			name: "found_partial_invalid",
			args: args{
				keysLength:       5,
				cacheMissIndexes: []int{0, 3},
				destSlicePtr:     &[]testResponseWithValid{{false}, {true}, {false}, {true}},
			},
			wantNewCacheMissIndexes: []int{0, 1, 3, 4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNewCacheMissIndexes, err := removeInvalidCache(tt.args.keysLength, tt.args.cacheMissIndexes, tt.args.destSlicePtr)
			if (err != nil) != tt.wantErr {
				t.Errorf("removeInvalidCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotNewCacheMissIndexes, tt.wantNewCacheMissIndexes) {
				t.Errorf("removeInvalidCache() = %v, want %v", gotNewCacheMissIndexes, tt.wantNewCacheMissIndexes)
			}
		})
	}
}
