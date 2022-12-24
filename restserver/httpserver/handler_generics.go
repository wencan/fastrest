//go:build go1.18
// +build go1.18

package httpserver

import "context"

// GenericsHandling 基于范型，将一个处理逻辑函数，转为Handling接口实现。
type GenericsHandling[REQUEST, RESPONSE any] func(ctx context.Context, request *REQUEST) (response *RESPONSE, err error)

// NewRequest 创建请求对象，返回请求对象的地址/指针。实现Handling接口。
func (handling GenericsHandling[REQUEST, RESPONSE]) NewRequest() interface{} {
	return new(REQUEST)
}

// Handle 处理请求，返回响应。实现Handling接口。
func (handling GenericsHandling[REQUEST, RESPONSE]) Handle(ctx context.Context, req interface{}) (resp interface{}, err error) {
	request := req.(*REQUEST)
	response, err := handling(ctx, request)
	if response != nil { // 警惕nil接口
		resp = response
	}
	return resp, err
}
