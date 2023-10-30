package restutils

import (
	"reflect"
	"testing"
)

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
