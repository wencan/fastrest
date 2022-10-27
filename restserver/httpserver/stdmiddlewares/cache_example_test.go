package stdmiddlewares

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/wencan/fastrest/restcache/lrucache"
)

func ExampleNewCacheMiddleware() {
	storage := lrucache.NewLRUCache(100, 10)
	ttlRange := [2]time.Duration{time.Minute * 4, time.Minute * 6}
	middleware := NewCacheMiddleware(storage, ttlRange, nil)

	s := httptest.NewServer(http.HandlerFunc(middleware(func(w http.ResponseWriter, r *http.Request) {
		requestURI := r.RequestURI
		name := r.URL.Query().Get("name")

		fmt.Printf("Received request. RequestURI: %s, Name: %s\n\n", requestURI, name)

		w.Header().Add("Name", name)
		w.Header().Add("RequestURI", requestURI)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hi, " + name))
	})))
	defer s.Close()

	client := s.Client()
	get := func(url string) {
		resp, err := client.Get(url)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		defer resp.Body.Close()
		fmt.Println("RequestURI:", resp.Header.Get("RequestURI"))
		fmt.Println("Name:", resp.Header.Get("Name"))
		data, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("Body:", string(data))
		fmt.Println()
	}

	get(s.URL + "/echo?name=Tom")
	get(s.URL + "/echo?name=Tom")
	get(s.URL + "/echo?name=Jerry")

	// Output:Received request. RequestURI: /echo?name=Tom, Name: Tom
	//
	// RequestURI: /echo?name=Tom
	// Name: Tom
	// Body: Hi, Tom
	//
	// RequestURI: /echo?name=Tom
	// Name: Tom
	// Body: Hi, Tom
	//
	// Received request. RequestURI: /echo?name=Jerry, Name: Jerry
	//
	// RequestURI: /echo?name=Jerry
	// Name: Jerry
	// Body: Hi, Jerry
}
