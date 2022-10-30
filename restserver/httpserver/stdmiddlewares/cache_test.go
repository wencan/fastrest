package stdmiddlewares

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/wencan/fastrest/restcache/mock_restcache"
	"github.com/wencan/fastrest/restutils"
)

func TestNewCacheMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		storageGetFunc  func(ctx context.Context, key string, valuePtr interface{}) (found bool, err error)
		storageSetFunc  func(ctx context.Context, key string, value interface{}, TTL time.Duration) error
		nextHandlerFunc http.HandlerFunc
		keyGenerator    RequestCacheKeyGenerator
		requestURI      string
	}
	type want struct {
		statusCode int
		headers    http.Header
		body       []byte
		error      bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "get_cached",
			args: args{
				storageGetFunc: func(ctx context.Context, key string, valuePtr interface{}) (found bool, err error) {
					resp := valuePtr.(*cacheableResponse)
					resp.StatusCode = http.StatusOK
					resp.Headers = http.Header{
						"Name": []string{"Tom"},
					}
					resp.Body = []byte("get_cached")
					return true, nil
				},
				nextHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotImplemented)
				},
				requestURI: "/get_cached?name=Tom",
			},
			want: want{
				statusCode: http.StatusOK,
				headers: http.Header{
					"Name":           []string{"Tom"},
					"Content-Length": []string{strconv.Itoa(len([]byte("get_cached")))},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
				},
				body: []byte("get_cached"),
			},
		},
		{
			name: "get_next",
			args: args{
				storageGetFunc: func(ctx context.Context, key string, valuePtr interface{}) (found bool, err error) {
					return false, nil
				},
				storageSetFunc: func(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
					return nil
				},
				nextHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("Name", "汤姆")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("测试"))
				},
				requestURI: "/get_next?name=汤姆",
			},
			want: want{
				statusCode: http.StatusOK,
				headers: http.Header{
					"Name":           []string{"汤姆"},
					"Content-Length": []string{strconv.Itoa(len([]byte("测试")))},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
				},
				body: []byte("测试"),
			},
		},
		{
			name: "get_cached_404",
			args: args{
				storageGetFunc: func(ctx context.Context, key string, valuePtr interface{}) (found bool, err error) {
					resp := valuePtr.(*cacheableResponse)
					resp.StatusCode = http.StatusNotFound
					resp.Headers = http.Header{
						"Name": []string{"Tom"},
					}
					return true, nil
				},
				nextHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotImplemented)
				},
				requestURI: "/get_cached_404?name=Tom",
			},
			want: want{
				statusCode: http.StatusNotFound,
				headers: http.Header{
					"Name":           []string{"Tom"},
					"Content-Length": []string{"0"},
				},
				body: []byte(""),
			},
		},
		{
			name: "get_next_404",
			args: args{
				storageGetFunc: func(ctx context.Context, key string, valuePtr interface{}) (found bool, err error) {
					return false, nil
				},
				storageSetFunc: func(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
					return nil
				},
				nextHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("Name", "汤姆")
					w.WriteHeader(http.StatusNotFound)
				},
				requestURI: "/get_next_404?name=汤姆",
			},
			want: want{
				statusCode: http.StatusNotFound,
				headers: http.Header{
					"Name":           []string{"汤姆"},
					"Content-Length": []string{"0"},
				},
				body: []byte(""),
			},
		},
		{
			name: "get_no_store",
			args: args{
				storageGetFunc: func(ctx context.Context, key string, valuePtr interface{}) (found bool, err error) {
					resp := valuePtr.(*cacheableResponse)
					resp.StatusCode = http.StatusOK
					resp.Headers = http.Header{
						"Name":          []string{"Tom"},
						"Cache-Control": []string{"no-Store"},
					}
					resp.GenerateTimestamp = time.Now().Unix()
					return true, nil
				},
				storageSetFunc: func(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
					return nil
				},
				nextHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Header().Add("From", "next")
					w.Write([]byte("welcome to next"))
				},
				requestURI: "/get_no_store",
			},
			want: want{
				statusCode: http.StatusOK,
				headers: http.Header{
					"From":           []string{"next"},
					"Content-Length": []string{strconv.Itoa(len([]byte("welcome to next")))},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
				},
				body: []byte("welcome to next"),
			},
		},
		{
			name: "get_maxage_valid",
			args: args{
				storageGetFunc: func(ctx context.Context, key string, valuePtr interface{}) (found bool, err error) {
					resp := valuePtr.(*cacheableResponse)
					resp.StatusCode = http.StatusOK
					resp.Headers = http.Header{
						"From":          []string{"Cache"},
						"Cache-Control": []string{"max-age=60"},
					}
					resp.Body = []byte("有效的缓存")
					resp.GenerateTimestamp = time.Now().Add(-time.Second * 50).Unix()
					return true, nil
				},
				storageSetFunc: func(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
					return nil
				},
				nextHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotImplemented)
				},
				requestURI: "/get_maxage_valid",
			},
			want: want{
				statusCode: http.StatusOK,
				headers: http.Header{
					"From":           []string{"Cache"},
					"Cache-Control":  []string{"max-age=60"},
					"Content-Length": []string{strconv.Itoa(len([]byte("有效的缓存")))},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
				},
				body: []byte("有效的缓存"),
			},
		},
		{
			name: "get_maxage_invalid",
			args: args{
				storageGetFunc: func(ctx context.Context, key string, valuePtr interface{}) (found bool, err error) {
					resp := valuePtr.(*cacheableResponse)
					resp.StatusCode = http.StatusOK
					resp.Headers = http.Header{
						"From":          []string{"Cache"},
						"Cache-Control": []string{"max-age=60"},
					}
					resp.Body = []byte("失效的缓存")
					resp.GenerateTimestamp = time.Now().Add(-time.Second * 80).Unix()
					return true, nil
				},
				storageSetFunc: func(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
					return nil
				},
				nextHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("From", "next")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Hi"))
				},
				requestURI: "/get_maxage_invalid",
			},
			want: want{
				statusCode: http.StatusOK,
				headers: http.Header{
					"From":           []string{"next"},
					"Content-Length": []string{strconv.Itoa(len([]byte("Hi")))},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
				},
				body: []byte("Hi"),
			},
		},
		{
			name: "get_expires_valid",
			args: args{
				storageGetFunc: func(ctx context.Context, key string, valuePtr interface{}) (found bool, err error) {
					resp := valuePtr.(*cacheableResponse)
					resp.StatusCode = http.StatusOK
					resp.Headers = http.Header{
						"From":    []string{"Cache"},
						"Expires": []string{time.Now().Format(time.RFC1123)},
					}
					resp.Body = []byte("未过期")
					resp.GenerateTimestamp = time.Now().Unix()
					return true, nil
				},
				storageSetFunc: func(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
					return nil
				},
				nextHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotImplemented)
				},
				requestURI: "/get_expires_valid",
			},
			want: want{
				statusCode: http.StatusOK,
				headers: http.Header{
					"From":           []string{"Cache"},
					"Content-Length": []string{strconv.Itoa(len([]byte("未过期")))},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
				},
				body: []byte("未过期"),
			},
		},
		{
			name: "get_expires_invalid",
			args: args{
				storageGetFunc: func(ctx context.Context, key string, valuePtr interface{}) (found bool, err error) {
					resp := valuePtr.(*cacheableResponse)
					resp.StatusCode = http.StatusOK
					resp.Headers = http.Header{
						"From":    []string{"Cache"},
						"Expires": []string{time.Now().Add(-time.Minute).Format(time.RFC1123)},
					}
					resp.Body = []byte("已过期")
					resp.GenerateTimestamp = time.Now().Add(-time.Minute).Unix()
					return true, nil
				},
				storageSetFunc: func(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
					return nil
				},
				nextHandlerFunc: func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("From", "next")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("你好"))
				},
				requestURI: "/get_expires_invalid",
			},
			want: want{
				statusCode: http.StatusOK,
				headers: http.Header{
					"From":           []string{"next"},
					"Content-Length": []string{strconv.Itoa(len([]byte("你好")))},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
				},
				body: []byte("你好"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mock_restcache.NewMockStorage(ctrl)
			if tt.args.storageGetFunc != nil {
				mockStorage.EXPECT().Get(gomock.Any(), gomock.AssignableToTypeOf(""), gomock.Any()).DoAndReturn(tt.args.storageGetFunc).MaxTimes(1)
			}
			if tt.args.storageSetFunc != nil {
				mockStorage.EXPECT().Set(gomock.Any(), gomock.AssignableToTypeOf(""), gomock.Any(), gomock.AssignableToTypeOf(time.Second)).DoAndReturn(tt.args.storageSetFunc).MaxTimes(1)
			}

			cacheMiddleware := NewCacheMiddleware(mockStorage, [2]time.Duration{time.Hour, time.Hour * 2}, tt.args.keyGenerator)
			handlerFunc := cacheMiddleware(tt.args.nextHandlerFunc)

			s := httptest.NewServer(http.HandlerFunc(handlerFunc))
			defer s.Close()

			client := s.Client()
			resp, err := client.Get(s.URL + tt.args.requestURI)
			if tt.want.error {
				assert.NotNil(t, err)
			} else {
				if assert.Nil(t, err) {
					defer resp.Body.Close()

					assert.Equal(t, tt.want.statusCode, resp.StatusCode)

					gotHeaders := make(http.Header)
					for key, values := range resp.Header {
						if restutils.StringSliceContains([]string{"Date", "Expires"}, key) { // 过滤掉日期时间
							continue
						}
						gotHeaders[key] = values
					}
					assert.Equal(t, tt.want.headers, gotHeaders)

					gotBody, err := io.ReadAll(resp.Body)
					if assert.Nil(t, err) {
						assert.Equal(t, tt.want.body, gotBody)
					}
				}
			}
		})
	}
}
