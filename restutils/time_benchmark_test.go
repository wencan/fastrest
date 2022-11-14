package restutils

import (
	"testing"
	"time"
)

func BenchmarkCoarseTimestamp(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			CoarseTimestamp()
		}
	})
}

func BenchmarkTimeNow(b *testing.B) {
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			time.Now().Second()
		}
	})
}
