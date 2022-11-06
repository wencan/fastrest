package httpserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func ExampleServer() {
	ctx := context.Background()

	s := NewServer(ctx, &http.Server{
		Addr: "127.0.0.1:28080",
		Handler: http.HandlerFunc(NewHandler(func(r *http.Request) (response interface{}, err error) {
			return struct {
				Echo string `json:"echo"`
			}{
				Echo: "Hello",
			}, nil
		})),
	})
	addr, err := s.Start(ctx) //  启动监听，开始服务。直至收到SIGTERM、SIGINT信号，或Stop被调用。
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Listen running at:", addr)

	resp, err := http.Get("http://" + addr + "/echo")
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
