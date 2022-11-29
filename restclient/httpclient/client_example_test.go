package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
)

func ExampleGet() {
	type Request struct {
		Greeting string `schema:"greeting"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		greeting := r.FormValue("greeting")

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf("{\"echo\":\"%s\"}\n", greeting)))
	}))
	defer s.Close()

	request := Request{
		Greeting: "Tom",
	}
	response := Response{}
	_ = Get(context.TODO(), &response, s.URL+"/test", request)
	fmt.Println(response.Echo)

	// Output: Tom
}

func ExamplePostJson() {
	type Request struct {
		Greeting string `json:"greeting"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := Request{}
		_ = json.NewDecoder(r.Body).Decode(&request)

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf("{\"echo\":\"%s\"}\n", request.Greeting)))
	}))
	defer s.Close()

	request := Request{
		Greeting: "Tom",
	}
	response := Response{}
	_ = PostJson(context.TODO(), &response, s.URL+"/test", request)
	fmt.Println(response.Echo)

	// Output: Tom
}

func ExampleClient() {
	// 先配置一个客户端
	client := DefaultClient
	client.NewRequestFunc = func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
		data, _ := io.ReadAll(body)
		request := struct {
			Data json.RawMessage `json:"data"`
		}{
			Data: data,
		}
		data, _ = json.Marshal(request)

		r, _ := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(data))
		r.Header.Set("Level", "Example")
		return r, nil
	}
	client.ReadResponseFunc = func(ctx context.Context, dest interface{}, response *http.Response) error {
		if response.StatusCode != http.StatusOK {
			return StatusCodeError(response.StatusCode, "upstream server error")
		}

		resp := struct {
			Data json.RawMessage `json:"data"`
		}{}
		_ = ReadResponseBody(ctx, &resp, response)

		_ = json.Unmarshal(resp.Data, dest)
		return nil
	}

	type Request struct {
		Greeting string `json:"greeting"`
	}
	type Response struct {
		Echo  string `json:"echo"`
		Level string `json:"level"`
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := struct {
			Data Request `json:"data"`
		}{}
		_ = json.NewDecoder(r.Body).Decode(&request)

		greeting := request.Data.Greeting
		level := r.Header.Get("Level")

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf("{\"data\": {\"echo\":\"%s\",\"level\":\"%s\"}}\n", greeting, level)))
	}))
	defer s.Close()

	request := Request{
		Greeting: "Tom",
	}
	response := Response{}
	_ = client.PostJson(context.TODO(), &response, s.URL+"/test", request)
	fmt.Println(response.Echo)
	fmt.Println(response.Level)

	// Output: Tom
	// Example
}
