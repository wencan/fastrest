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

func ExampleNewHandlerFactory() {
	factory := NewHandlerFactory(&HandlerFactoryConfig{
		RequestInterceptor: func(r *http.Request) (overwriteRequest *http.Request, err error) {
			fmt.Println(r.RequestURI)
			return r, nil
		},
		ResponseInterceptor: func(ctx context.Context, response interface{}, err error) (overwriteResponse interface{}, overwriteErr error) {
			return struct {
				Data interface{} `json:"data"`
			}{
				Data: response,
			}, nil
		},
	})
	handler := factory(func(r *http.Request) (response interface{}, err error) {
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
