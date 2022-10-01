# fastrest
Go语言RESTful服务通用组件  
Restful服务公共组件库，目的为帮忙快速开发服务程序，尽可能省去与业务无关的重复代码。  
可以只使用个别组件，也可以组合起来当框架用。

## 示例

创建一个HTTP Handler。
```go
	var handler http.HandlerFunc = NewHandler(func(r *http.Request) (response interface{}, err error) {
		req := struct {
			Greeting string `schema:"greeting"`
		}{}
        // parse query
		err = ReadRequest(r.Context(), &req, r)
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