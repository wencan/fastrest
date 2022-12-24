//go:build go1.18
// +build go1.18

package httpserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

func ExampleNewHandler() {
	type Request struct {
		Greeting string `schema:"greeting" validate:"required"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	f := func(ctx context.Context, req *Request) (resp *Response, err error) {
		return &Response{
			Echo: req.Greeting,
		}, nil
	}
	var handler http.HandlerFunc = NewHandler(GenericsHandling[Request, Response](f))

	s := httptest.NewServer(handler)
	defer s.Close()

	client := s.Client()
	resp, err := client.Get(s.URL + "/echo?greeting=hello")
	if err != nil {
		fmt.Println(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("status code: %d", resp.StatusCode)
		return
	}

	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(data))
	// Output: {"echo":"hello"}
}

func ExampleNewHandler_withMiddleware() {
	type Request struct {
		Greeting string `schema:"greeting" validate:"required"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	middleware := ChainHandlerMiddlewares(func(next HandleFunc) HandleFunc {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			response, err = next(ctx, request)
			if err != nil {
				return nil, err
			}
			return struct {
				Data interface{} `json:"data"`
			}{
				Data: response,
			}, nil
		}
	}, RecoveryMiddleware)

	factory := DefaultHandlerFactory
	factory.Middleware = middleware
	f := func(ctx context.Context, req *Request) (resp *Response, err error) {
		return &Response{
			Echo: req.Greeting,
		}, nil
	}
	var handler http.HandlerFunc = factory.NewHandler(GenericsHandling[Request, Response](f))

	s := httptest.NewServer(handler)
	defer s.Close()

	client := s.Client()
	resp, err := client.Get(s.URL + "/echo?greeting=hello")
	if err != nil {
		fmt.Println(err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("status code: %d", resp.StatusCode)
		return
	}

	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(data))

	// Output: {"data":{"echo":"hello"}}
}
