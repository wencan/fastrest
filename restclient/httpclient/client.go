package httpclient

import (
	"context"
	"net/http"

	"github.com/wencan/fastrest/restcodecs/restmime"
)

// DefaultClient 默认的客户端。可覆盖。
var DefaultClient = Client{
	NewRequestFunc:   http.NewRequestWithContext,
	DefaultAccept:    "*/*",
	DoFunc:           http.DefaultClient.Do,
	ReadResponseFunc: ReadResponse,
	UserAgent:        "fastrest-http-client/1.1",
}

// Client 带配置的客户端。
type Client struct {
	// NewRequestFunc 创建请求的函数。默认为：http.NewRequestWithContext。
	NewRequestFunc NewRequestFunc

	// DefaultAccept 请求Header的默认Accept值。默认为：*/*。
	DefaultAccept string

	// DoFunc 发送请求，解析响应的函数。默认为http.DefaultClient.Do。
	DoFunc func(req *http.Request) (*http.Response, error)

	// ReadResponseFunc 解析响应的函数。默认为：ReadResponseBody
	ReadResponseFunc ReadResponseFunc

	// UserAgent 请求Header的User-Agent值。默认fastrest-http-client/1.1。
	UserAgent string
}

// Do 发送请求，解析响应到对象。
func (client Client) Do(ctx context.Context, dest interface{}, r *http.Request) error {
	r.Header.Set("User-Agent", client.UserAgent)
	if client.DefaultAccept != "" && r.Header.Get("Accept") == "" {
		r.Header.Set("Accept", client.DefaultAccept)
	}

	response, err := client.DoFunc(r)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = client.ReadResponseFunc(ctx, dest, response)
	if err != nil {
		return err
	}

	return nil
}

// NewRequestWithBody 创建带请求body的http.Request。
// body可以为nil。
func (client Client) NewRequestWithBody(ctx context.Context, method, url, contentType string, body interface{}) (*http.Request, error) {
	var r *http.Request
	var err error
	if body == nil {
		r, err = client.NewRequestFunc(ctx, method, url, nil)
	} else {
		r, err = newRequestWithBody(ctx, method, url, contentType, body, client.NewRequestFunc)
	}
	if err != nil {
		return nil, err
	}

	return r, nil
}

// DoPost 发送post请求，解析响应到对象。
func (client Client) DoPost(ctx context.Context, dest interface{}, method, url, contentType string, body interface{}) error {
	r, err := client.NewRequestWithBody(ctx, method, url, contentType, body)
	if err != nil {
		return err
	}

	return client.Do(ctx, dest, r)
}

// NewRequestWithQuery 创建一个带查询字符串的http.Request。
// query可以为nil、url.Values，带schema标签的结构体。
func (client Client) NewRequestWithQuery(ctx context.Context, method, url string, query interface{}) (*http.Request, error) {
	return newRequestWithQuery(ctx, method, url, query, client.NewRequestFunc)
}

// Get 发送一个Get查询请求。query可以为nil、url.Values、带schema标签的结构体对象。
func (client Client) Get(ctx context.Context, dest interface{}, url string, query interface{}) error {
	r, err := client.NewRequestWithQuery(ctx, http.MethodGet, url, query)
	if err != nil {
		return err
	}

	return client.Do(ctx, dest, r)
}

// Post 发送一个Post请求。dest为接收响应的对象地址，可以为nil。
func (client Client) Post(ctx context.Context, dest interface{}, url string, contentType string, body interface{}) error {
	return client.DoPost(ctx, dest, http.MethodPost, url, contentType, body)
}

// PostJson 发送一个Post请求。请求实体为Json。dest为接收响应的对象地址，可以为nil。
func (client Client) PostJson(ctx context.Context, dest interface{}, url string, body interface{}) error {
	return client.DoPost(ctx, dest, http.MethodPost, url, restmime.MimeTypeJson, body)
}

// PostForm 发送一个Post请求。请求实体为form。dest为接收响应的对象地址，可以为nil。
func (client Client) PostForm(ctx context.Context, dest interface{}, url string, body interface{}) error {
	return client.DoPost(ctx, dest, http.MethodPost, url, restmime.MimeTypeForm, body)
}

// Get 基于DefaultClient，发送一个Get查询请求。query可以为nil、url.Values、带schema标签的结构体对象。
func Get(ctx context.Context, dest interface{}, url string, query interface{}) error {
	return DefaultClient.Get(ctx, dest, url, query)
}

// Post 基于DefaultClient，发送一个Post请求。dest为接收响应的对象地址，可以为nil。
func Post(ctx context.Context, dest interface{}, url string, contentType string, body interface{}) error {
	return DefaultClient.Post(ctx, dest, url, contentType, body)
}

// PostJson 基于DefaultClient，发送一个Post请求。请求实体为Json。dest为接收响应的对象地址，可以为nil。
func PostJson(ctx context.Context, dest interface{}, url string, body interface{}) error {
	return DefaultClient.PostJson(ctx, dest, url, body)
}

// PostForm 基于DefaultClient，发送一个Post请求。请求实体为form。dest为接收响应的对象地址，可以为nil。
func PostForm(ctx context.Context, dest interface{}, url string, contentType string, body interface{}) error {
	return DefaultClient.PostForm(ctx, dest, url, body)
}
