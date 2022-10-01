package restmime

import (
	"reflect"
	"testing"
)

func TestContentTypeNames(t *testing.T) {
	type args struct {
		contentType string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "simple",
			args: args{
				contentType: "application/xhtml+xml",
			},
			want: []string{"application/xhtml+xml"},
		},
		{
			name: "with_args",
			args: args{
				contentType: "application/xml;q=0.9",
			},
			want: []string{"application/xml"},
		},
		{
			name: "multi",
			args: args{
				contentType: "text/html, application/xhtml+xml, application/xml;q=0.9, image/webp, */*;q=0.8",
			},
			want: []string{"text/html", "application/xhtml+xml", "application/xml", "image/webp", "*/*"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContentTypeNames(tt.args.contentType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContentTypeNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContentTypeName(t *testing.T) {
	type args struct {
		contentType string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple",
			args: args{
				contentType: "application/xhtml+xml",
			},
			want: "application/xhtml+xml",
		},
		{
			name: "with_args",
			args: args{
				contentType: "application/xml;q=0.9",
			},
			want: "application/xml",
		},
		{
			name: "multi",
			args: args{
				contentType: "text/html, application/xhtml+xml, application/xml;q=0.9, image/webp, */*;q=0.8",
			},
			want: "text/html",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ContentTypeName(tt.args.contentType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ContentTypeName() = %s, want %s", got, tt.want)
			}
		})
	}
}
