package stdmiddlewares

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wencan/fastrest/restcache"
)

// RequestCacheKeyGenerator 根据http.Request生成缓存key。如果返回空字符串，表示不使用缓存。
type RequestCacheKeyGenerator func(r *http.Request) string

// DefaultRequestCacheKeyGenerator 默认的http.Request缓存key生成器。可覆盖。
var DefaultRequestCacheKeyGenerator = func(r *http.Request) string {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
	default:
		return ""
	}

	if strings.EqualFold(r.Header.Get("Cache-Control"), "no-Cache") {
		return ""
	}

	if r.RequestURI != "" {
		// 支持单元测试。如果Request不是解析得到，而是NewRequest得到，RequestURI为空。
		return fmt.Sprintf("%s:%s:%s", r.Method, r.Host, r.RequestURI)
	} else {
		return fmt.Sprintf("%s:%s", r.Method, r.URL.String())
	}
}

type cacheableResponse struct {
	StatusCode int `json:"status_code" msgpack:"status_code"`

	Headers http.Header `json:"headers" msgpack:"headers"`

	Body []byte `json:"body" msgpack:"body"`

	// GenerateTimestamp 响应生成时的时间辍。用于判断缓存是否有效。
	GenerateTimestamp int64 `json:"generate_ts" msgpack:"generate_ts"`
}

// Header 实现http.ResponseWriter接口。
func (resp *cacheableResponse) Header() http.Header {
	if resp.Headers == nil {
		resp.Headers = make(http.Header)
	}
	return resp.Headers
}

// Write 实现http.ResponseWriter接口。
func (resp *cacheableResponse) Write(p []byte) (int, error) {
	resp.Body = append(resp.Body, p...) // 后面优化
	return len(p), nil
}

// WriteHeader 实现http.ResponseWriter接口。
func (resp *cacheableResponse) WriteHeader(statusCode int) {
	resp.StatusCode = statusCode
}

// IsValidCache 实现restcache的Validatable接口。
func (resp cacheableResponse) IsValidCache() bool {
	cacheCtrl := resp.Headers.Get("Cache-Control")

	if strings.EqualFold(cacheCtrl, "no-Store") {
		return false
	}

	// 下面根据时间判断
	if resp.GenerateTimestamp <= 0 {
		return true
	}

	if strings.HasPrefix(cacheCtrl, "max-age=") {
		maxAge, _ := strconv.Atoi(strings.ReplaceAll(cacheCtrl, "max-age=", ""))
		age := sinceUnix(resp.GenerateTimestamp)
		return maxAge >= int(age)
	}
	if strings.HasPrefix(cacheCtrl, "s-maxage=") {
		maxAge, _ := strconv.Atoi(strings.ReplaceAll(cacheCtrl, "s-maxage=", ""))
		age := sinceUnix(resp.GenerateTimestamp)
		return maxAge >= int(age)
	}

	expires := resp.Headers.Get("Expires")
	if expires == "0" {
		return true
	}
	expiresTime, _ := time.Parse(time.RFC1123, expires)
	if !expiresTime.IsZero() {
		return expiresTime.Unix() >= getNowUnix()
	}

	return true
}

// Reset 实现restcache的Resetable接口。
// 当对象被污染后，reset内容。
func (resp *cacheableResponse) Reset() {
	resp.StatusCode = 0
	resp.Headers = nil
	resp.Body = nil
}

// Apply 输出。
func (resp *cacheableResponse) Apply(w http.ResponseWriter) {
	for key, values := range resp.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	if resp.Body != nil {
		w.Write(resp.Body)
	}
}

var cacheableResponsePool = sync.Pool{New: func() interface{} {
	return &cacheableResponse{
		Headers: make(http.Header),
	}
}}

type handlerQueryArgs struct {
	request *http.Request

	next http.HandlerFunc
}

// NewCacheMiddleware 创建http.Handler的缓存中间件。
// storage 为缓存存储器。
// ttlRange 为缓存生存时间区间。
// keyGenerator 为缓存key生成器。默认为：DefaultRequestCacheKeyGenerator。
// 支持简单的常用的HTTP缓存控制。
func NewCacheMiddleware(
	storage restcache.Storage,
	ttlRange [2]time.Duration,
	keyGenerator RequestCacheKeyGenerator,
) func(next http.HandlerFunc) http.HandlerFunc {
	if keyGenerator == nil {
		keyGenerator = DefaultRequestCacheKeyGenerator
	}

	// 如果没命中，要执行的过程
	query := func(ctx context.Context, destPtr, args interface{}) (found bool, err error) {
		w := destPtr.(*cacheableResponse)

		queryArgs := args.(*handlerQueryArgs)
		r := queryArgs.request
		next := queryArgs.next

		next(w, r)
		w.GenerateTimestamp = time.Now().Unix() // 用于支持缓存控制
		return true, nil
	}

	// 缓存中间件
	caching := restcache.Caching{Storage: storage, Query: query, TTLRange: ttlRange, SentinelTTL: time.Second}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			key := keyGenerator(r)
			if key == "" { // 不需要或者不支持缓存
				next(w, r)
				return
			}

			// 执行缓存逻辑
			// 如果没命中缓存，由缓存中间件去执行next
			resp := cacheableResponse{}
			args := &handlerQueryArgs{request: r, next: next}
			found, err := caching.Get(r.Context(), &resp, key, args)
			if !found || err != nil { // 这里不应该返回not found，或者err != nil
				fmt.Fprintf(os.Stderr, "Error in cache middleware")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// 输出
			resp.Apply(w)
		}
	}
}
