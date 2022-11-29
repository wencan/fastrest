# fastrest

[![Go Reference](https://pkg.go.dev/badge/github.com/wencan/fastrest)](https://pkg.go.dev/github.com/wencan/fastrest)  


Go语言RESTful服务通用组件库  
Restful服务公共组件库，目的为帮忙快速开发服务程序，尽可能省去与业务无关的重复代码。  
可以只使用个别组件，也可以组合起来当“框架”用。


## fastrest与常说的“框架”的区别
* fastrest是一些辅助性质的函数/结构的集合。
* fastrest不造重复的轮子。
* fastrest力求足够的开放。

## 目录  
<table>
    <tr>
        <th>包</th><th>结构体/方法</th><th>作用</th><th>说明</th>
    </tr>
    <tr>
        <td><a href="https://pkg.go.dev/github.com/wencan/fastrest/restserver/httpserver">restserver/httpserver</a></td><td></td><td>http服务组件</td><td>一套http服务的辅助组件，需要组合<a href="https://pkg.go.dev/net/http">http</a>、<a href="https://pkg.go.dev/net/http#ServeMux">multiplexer</a>一起使用</td>
    </tr>
    <tr>
        <td><a href="https://pkg.go.dev/github.com/wencan/fastrest/restserver/httpserver/stdmiddlewares">restserver/httpserver/stdmiddlewares</a></td><td></td><td>http中间件</td><td>一个http的缓存中间件，支持简单的常见的缓存控制策略</td>
    </tr>
    <tr>
        <td><a href="https://pkg.go.dev/github.com/wencan/fastrest/restclient/httpclient">restclient/httpclient</a></td><td></td><td>http客户端组件</td><td>一套http客户端的辅助组件</td>
    </tr>
    <tr>
        <td rowspan="2">restcache</td><td><a href="https://pkg.go.dev/github.com/wencan/fastrest/restcache#Caching">Caching</a></td><td>单个数据的缓存中间件</td><td rowspan="2">缓存流程的胶水逻辑。<br>基于<a href="https://pkg.go.dev/github.com/wencan/gox/xsync/sentinel#SentinelGroup">SentinelGroup</a>解决缓存实效风暴问题。<br>简单介绍见<a href="https://blog.wencan.org/2022/10/17/restcache/">这里</a>。</td>
    </tr>
    <tr>
        <td><a href="https://pkg.go.dev/github.com/wencan/fastrest/restcache#MCaching">MCaching</a></td><td>批量数据的缓存中间件</td>
    </tr>
    <tr>
        <td>restcache/lrucache</td><td><a href="https://pkg.go.dev/github.com/wencan/fastrest/restcache/lrucache#LRUCache">LRUCache</a></td><td>LRU缓存存储</td><td>实现了restcache的缓存存储接口。<br>基于<a href="https://pkg.go.dev/github.com/wencan/gox/xsync#LRUMap">LRUMap</a>实现。</td>
    </tr>
</table>

## 示例

### restserver/httpserver：启动一个HTTP Server
```go
s := NewServer(ctx, &http.Server{
    Addr: "127.0.0.1:28080",
    Handler: ...,
})
addr, err := s.Start(ctx) //  启动监听，开始服务。直至收到SIGTERM、SIGINT信号，或Stop被调用。
fmt.Println("Listen running at:", addr)

s.Stop(ctx)   // 结束监听
err = s.Wait(ctx) // 等待处理完
```

### restserver/httpserver: 创建一个HTTP Handler
```go
type Request struct {
    Greeting string `schema:"greeting" json:"greeting" validate:"required"`
}
type Response struct {
    Echo string `json:"echo"`
}

var handler http.HandlerFunc = NewHandler(func(r *http.Request) (response interface{}, err error) {
    var req Request
    // parse and validate query
    err = ReadValidateRequest(r.Context(), &req, r)
    if err != nil {
        return nil, err
    }

    // do things

    // output json body
    return Response{
        Echo: req.Greeting,
    }, nil
})
```

或者：
```go
type Request struct {
    Greeting string `schema:"greeting" json:"greeting" validate:"required"`
}
type Response struct {
    Echo string `json:"echo"`
}

// NewHandler + NewHandlerFunc 可以将一个gRPC服务方法，转为http.HandlerFunc。
var handler http.HandlerFunc = NewHandler(NewHandlerFunc(func(ctx context.Context, req *Request) (resp *Response, err error) {
    // do things

    return &Response{
        Echo: req.Greeting,
    }, nil
}, ReadValidateRequest))
```

### fastrest/restserver/httpserver/stdmiddlewares：HTTP缓存中间件
```go
storage := lrucache.NewLRUCache(10000, 10)    // 一般也可能是redis客户端
ttlRange := [2]time.Duration{time.Minute * 4, time.Minute * 6}
cacheMiddleware := NewCacheMiddleware(storage, ttlRange, nil)

var handler http.HandlerFunc = cacheMiddleware(func(w http.ResponseWriter, r *http.Request) {
    // 缓存未命中，才会执行这里
    // ... ...
})
```

### restclient/httpclient: 客户端Get请求
```go
	type Request struct {
		Greeting string `schema:"greeting"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	request := Request{Greeting: "Hi"}
	response := Response{}
	_ = Get(context.TODO(), &response, s.URL+"/path", request)
```


### restclient/httpclient: 客户端PostJson请求
```go
	type Request struct {
		Greeting string `json:"greeting"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	request := Request{Greeting: "Hi"}
	response := Response{}
	_ = PostJson(context.TODO(), &response, s.URL+"/path", request)
```


### restcache: 单个数据的缓存
```go
query := func(ctx context.Context, destPtr interface{}, args interface{}) (found bool, err error) {
    // 如果Storage没找到，这里提供
    // 一般是从持久化数据库、服务接口查询
    return true, nil
}

caching := Caching{
    Storage:     lrucache.NewLRUCache(10000, 10), // 一般是redis实现。透明处理业务数据。
    Query:       query,
    TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
    SentinelTTL: time.Second,
}

var resp Response
var key = "key"
var args = ?    // 给query函数的参数
found, err := caching.Get(context.TODO(), &resp, key, args)
``` 

### restcache: 一批数据的缓存
```go
query := func(ctx context.Context, destSlicePtr interface{}, argsSlice interface{}) (missIndexes []int, err error) {
    // 如果Storage没找到，这里提供
    // 一般是从持久化数据库、服务接口查询
    // destSlicePtr为接收结果的切片的指针，argsSlice为入参切片
    // missIndexes是没查到的下标
    return missIndexes, nil
}

mcaching := MCaching{
    MStorage:    lrucache.NewLRUCache(10000, 10), // 一般是redis实现
    MQuery:      query,
    TTLRange:    [2]time.Duration{time.Minute * 4, time.Minute * 6},
    SentinelTTL: time.Second,
}

var keys = []string{......}
var args = ......    // 给query函数的参数切片
var resp []*response
missIndexes, err := mcaching.MGet(context.TODO(), &resp, keys, args)

for _, missIndex := range missIndexes {
    // not found: keys[missIndex]
}
```