//go:build go1.18
// +build go1.18

package httpserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func ExampleServer() {
	type Request struct {
		Greeting string `schema:"greeting" validate:"required"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}

	ctx := context.Background()

	s := NewServer(ctx, &http.Server{
		Addr: "127.0.0.1:28080",
		Handler: http.HandlerFunc(NewHandler(GenericsHandling[Request, Response](func(ctx context.Context, req *Request) (*Response, error) {
			return &Response{
				Echo: req.Greeting,
			}, nil
		}))),
	})
	addr, err := s.Start(ctx) //  启动监听，开始服务。直至收到SIGTERM、SIGINT信号，或Stop被调用。
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Listen running at:", addr)

	resp, err := http.Get("http://" + addr + "/echo?greeting=Hello")
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

	s.Stop(ctx)       // 结束监听
	err = s.Wait(ctx) // 等待处理完
	if err != nil {
		fmt.Println(err)
		return
	}

	// Output: Listen running at: 127.0.0.1:28080
	// {"echo":"Hello"}
}
