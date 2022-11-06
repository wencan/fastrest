package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/wencan/fastrest/restcache/lrucache"
	"github.com/wencan/fastrest/restserver/httpserver"
	"github.com/wencan/fastrest/restserver/httpserver/stdmiddlewares"
)

var addr = flag.String("addr", "127.0.0.1:8080", "listen address")

func main() {
	flag.Parse()

	lrustorage := lrucache.NewLRUCache(1000, 10)
	ttlRange := [2]time.Duration{time.Hour * 23, time.Hour * 26}
	cacheMiddleware := stdmiddlewares.NewCacheMiddleware(lrustorage, ttlRange, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/proxy/", cacheMiddleware(func(w http.ResponseWriter, r *http.Request) {
		targetURL := r.URL.Query().Get("target_url")
		if r.Method != http.MethodGet || targetURL == "" {
			log.Println("invalid request:", r.Method, r.RequestURI)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid request"))
			return
		}

		log.Println("GET", targetURL)
		response, err := http.Get(targetURL)
		if err != nil {
			log.Println(targetURL, "reply:", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(err.Error()))
			return
		}
		defer response.Body.Close()

		for key, headers := range response.Header {
			for _, header := range headers {
				w.Header().Add(key, header)
			}
		}
		w.WriteHeader(response.StatusCode)
		_, err = io.Copy(w, response.Body)
		if err != nil {
			log.Println("write responser error:", err)
		}
	}))

	ctx := context.Background()
	var srv = httpserver.NewServer(ctx, &http.Server{
		Addr:    *addr,
		Handler: mux,
	})
	addr, err := srv.Start(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Listen running at:", addr)

	err = srv.Wait(ctx)
	if err != nil {
		log.Fatalln(err)
	}
}
