package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/wencan/fastrest/restcache/lrucache"
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

	var srv = http.Server{
		Addr:    *addr,
		Handler: mux,
	}
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("Server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatalln("Listen:", err)
	}
	log.Println("Listen running at:", srv.Addr)
	err = srv.Serve(ln)
	if err != nil {
		if err == http.ErrServerClosed {
			log.Println("Server closed")
		} else {
			log.Fatalln("Serve:", err)
		}
	}
	<-idleConnsClosed
}
