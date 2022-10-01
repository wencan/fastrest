package restutils

import "testing"

func TestIsGhostInterface(t *testing.T) {
	type NilInterface interface{}
	var nilInterface NilInterface = (*int)(nil)
	var nonNilInterface NilInterface = struct{}{}

	type args struct {
		i interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test_nil_interface",
			args: args{
				i: nil,
			},
			want: true,
		},
		{
			name: "test_target_nil",
			args: args{
				i: nilInterface,
			},
			want: true,
		},
		{
			name: "test_target_non_nil",
			args: args{
				i: nonNilInterface,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGhostInterface(tt.args.i); got != tt.want {
				t.Errorf("IsGhostInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}
