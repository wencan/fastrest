package restcache

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_getTTL_Concurrently(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < 10000; j++ {
				startDuratin := time.Second * time.Duration(rand.Int63n(10000))
				endDuratin := startDuratin + time.Second*time.Duration(rand.Int63n(10000))
				ttl := getTTL([2]time.Duration{startDuratin, endDuratin})
				assert.True(t, ttl >= startDuratin, "ttl: %v, range: [%v: %v]", ttl, startDuratin, endDuratin)
				assert.True(t, ttl <= endDuratin, "ttl: %v, range: [%v: %v]", ttl, startDuratin, endDuratin)
			}
		}()
	}
	wg.Wait()
}
