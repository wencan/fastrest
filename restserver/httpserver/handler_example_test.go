package httpserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

func ExampleNewHandler() {
	var handler http.HandlerFunc = NewHandler(func(r *http.Request) (response interface{}, err error) {
		req := struct {
			Greeting string `schema:"greeting" validate:"required"`
		}{}
		// 解析并验证请求
		err = ReadValidateRequest(r.Context(), &req, r)
		if err != nil {
			return nil, err
		}

		return struct {
			Echo string `json:"echo"`
		}{
			Echo: req.Greeting,
		}, nil
	})

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

func ExampleNewHandlerFunc() {
	type Request struct {
		Greeting string `schema:"greeting" validate:"required"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}
	var handler http.HandlerFunc = NewReflectHandler(func(ctx context.Context, req *Request) (resp Response, err error) {
		return Response{
			Echo: req.Greeting,
		}, nil
	})

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
	middleware := ChainHandlerMiddlewares(func(next HandlerFunc) HandlerFunc {
		return func(r *http.Request) (response interface{}, err error) {
			fmt.Println(r.RequestURI)
			response, err = next(r)
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

	handler := NewHandler(middleware(func(r *http.Request) (response interface{}, err error) {
		req := struct {
			Greeting string `schema:"greeting"`
		}{}
		err = ReadRequest(r.Context(), &req, r)
		if err != nil {
			return nil, err
		}

		return struct {
			Echo string `json:"echo"`
		}{
			Echo: req.Greeting,
		}, nil
	}))

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

	// Output: /echo?greeting=hello
	// {"data":{"echo":"hello"}}
}

func ExampleNewHandlerFunc_withMiddleware() {
	type Request struct {
		Greeting string `schema:"greeting" validate:"required"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	middleware := ChainHandlerMiddlewares(func(next HandlerFunc) HandlerFunc {
		return func(r *http.Request) (response interface{}, err error) {
			fmt.Println(r.RequestURI)
			response, err = next(r)
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
	handler := NewHandler(middleware(NewHandlerFunc(func(ctx context.Context, req *Request) (resp Response, err error) {
		return Response{
			Echo: req.Greeting,
		}, nil
	}, nil)))

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

	// Output: /echo?greeting=hello
	// {"data":{"echo":"hello"}}
}

func ExampleHandlerFactory() {
	type Request struct {
		Greeting string `schema:"greeting" validate:"required"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	factory := DefaultHandlerFactory
	factory.ReadRequestFunc = ReadValidateRequest
	factory.Middleware = ChainHandlerMiddlewares(func(next HandlerFunc) HandlerFunc {
		return func(r *http.Request) (response interface{}, err error) {
			fmt.Println(r.RequestURI)
			response, err = next(r)
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

	handler := factory.NewReflectHandler(func(ctx context.Context, req *Request) (response *Response, err error) {
		return &Response{
			Echo: req.Greeting,
		}, nil
	})

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

	// Output: /echo?greeting=hello
	// {"data":{"echo":"hello"}}
}
