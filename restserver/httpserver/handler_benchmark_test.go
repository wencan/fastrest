//go:build go1.18
// +build go1.18

package httpserver

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkNewHandler(b *testing.B) {
	type Request struct {
		Greeting string `schema:"greeting" validate:"required"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}
	var handler http.HandlerFunc = NewHandler(GenericsHandling[Request, Response](func(ctx context.Context, req *Request) (*Response, error) {
		return &Response{
			Echo: req.Greeting,
		}, nil
	}))

	s := httptest.NewServer(handler)
	defer s.Close()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			client := s.Client()
			resp, err := client.Get(s.URL + "/echo?greeting=hello")
			if err != nil {
				b.Fatal(err)
				return
			}
			if resp.StatusCode != http.StatusOK {
				b.Fatalf("status code: %d", resp.StatusCode)
				return
			}

			_, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				b.Fatal(err)
				return
			}
		}
	})
}

func BenchmarkNewHandler_withoutServe(b *testing.B) {
	type Request struct {
		Greeting string `schema:"greeting" validate:"required"`
	}
	type Response struct {
		Echo string `json:"echo"`
	}
	var handler http.HandlerFunc = NewHandler(GenericsHandling[Request, Response](func(ctx context.Context, req *Request) (*Response, error) {
		return &Response{
			Echo: req.Greeting,
		}, nil
	}))

	r, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "/echo?greeting=hello", nil)
	if err != nil {
		b.Fatal(err)
		return
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			w := httptest.NewRecorder()
			handler(w, r)
		}
	})
}
