package restcache

import (
	"math/rand"
	"reflect"
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

func TestHitIndexes(t *testing.T) {
	tests := []struct {
		name        string
		collection  []int
		missIndexes []int
		want        []int
	}{
		{
			name:        "all_hit",
			collection:  []int{1, 2, 3, 4, 5, 6},
			missIndexes: nil,
			want:        []int{0, 1, 2, 3, 4, 5},
		},
		{
			name:        "no_hit",
			collection:  []int{1, 2, 3, 4, 5, 6},
			missIndexes: []int{0, 1, 2, 3, 4, 5},
			want:        nil,
		},
		{
			name:        "miss",
			collection:  []int{1, 2, 3, 4, 5, 6},
			missIndexes: []int{0, 2, 4, 5},
			want:        []int{1, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HitIndexes(tt.collection, tt.missIndexes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HitIndexes() = %v, want %v", got, tt.want)
			}
		})
	}
}
