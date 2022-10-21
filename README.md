# fastrest

[![Go Reference](https://pkg.go.dev/badge/github.com/wencan/fastrest)](https://pkg.go.dev/github.com/wencan/fastrest)  


Go语言RESTful服务通用组件  
Restful服务公共组件库，目的为帮忙快速开发服务程序，尽可能省去与业务无关的重复代码。  
可以只使用个别组件，也可以组合起来当框架用。

## 目录  
<table>
    <tr>
        <th>包</th><th>结构体/方法</th><th>作用</th><th>说明</th>
    </tr>
    <tr>
        <td><a href="https://pkg.go.dev/github.com/wencan/fastrest/restserver/httpserver">restserver/httpserver</a></td><td></td><td>http服务组件</td><td>未完成</td>
    </tr>
    <tr>
        <td rowspan="2">restcache</td><td><a href="https://pkg.go.dev/github.com/wencan/fastrest/restcache#Caching">Caching</a></td><td>单个数据的缓存</td><td rowspan="2">基于<a href="https://pkg.go.dev/github.com/wencan/gox/xsync/sentinel#SentinelGroup">SentinelGroup</a>解决缓存实效风暴问题。简单介绍见<a href="https://blog.wencan.org/2022/10/17/restcache/">这里</a>。</td>
    </tr>
    <tr>
        <td><a href="https://pkg.go.dev/github.com/wencan/fastrest/restcache#MCaching">MCaching</a></td><td>批量数据的缓存</td>
    </tr>
    <tr>
        <td>restcache/lrucache</td><td><a href="https://pkg.go.dev/github.com/wencan/fastrest/restcache/lrucache#LRUCache">LRUCache</a></td><td>LRU缓存存储</td><td>基于<a href="https://pkg.go.dev/github.com/wencan/gox/xsync#LRUMap">LRUMap</a>实现</td>
    </tr>
</table>

## 示例

### restserver/httpserver: 创建一个HTTP Handler
```go
var handler http.HandlerFunc = NewHandler(func(r *http.Request) (response interface{}, err error) {
    req := struct {
        Greeting string `schema:"greeting" validate:"required"`
    }{}
    // parse and validate query
    err = ReadValidateRequest(r.Context(), &req, r)
    if err != nil {
        return nil, err
    }

    // do things

    // output json body
    return struct {
        Echo string `json:"echo"`
    }{
        Echo: req.Greeting,
    }, nil
})
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