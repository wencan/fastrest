package restutils

import (
	"reflect"
	"testing"
)

func TestParseDataUrls(t *testing.T) {
	tests := []struct {
		name          string
		rawURL        string
		wantMediaType string
		wantIsBase64  bool
		wantData      []byte
		wantErr       bool
	}{
		{
			name:     "data:,Hello%2C%20World%21",
			rawURL:   "data:,Hello%2C%20World%21",
			wantData: []byte("Hello, World!"),
		},
		{
			name:          "data:text/plain;base64,SGVsbG8sIFdvcmxkIQ==",
			rawURL:        "data:text/plain;base64,SGVsbG8sIFdvcmxkIQ==",
			wantMediaType: "text/plain",
			wantIsBase64:  true,
			wantData:      []byte("Hello, World!"),
		},
		{
			name:          "data:text/plain;base64,5oiR5piv5rWL6K+V5pWw5o2u",
			rawURL:        "data:text/plain;base64,5oiR5piv5rWL6K+V5pWw5o2u",
			wantMediaType: "text/plain",
			wantIsBase64:  true,
			wantData:      []byte("我是测试数据"),
		},
		{
			name:          "data:text/html,%3Ch1%3EHello%2C%20World%21%3C%2Fh1%3E",
			rawURL:        "data:text/html,%3Ch1%3EHello%2C%20World%21%3C%2Fh1%3E",
			wantMediaType: "text/html",
			wantData:      []byte("<h1>Hello, World!</h1>"),
		},
		{
			name:          "data:text/html,%3Cscript%3Ealert%28%27hi%27%29%3B%3C%2Fscript%3E",
			rawURL:        "data:text/html,%3Cscript%3Ealert%28%27hi%27%29%3B%3C%2Fscript%3E",
			wantMediaType: "text/html",
			wantData:      []byte("<script>alert('hi');</script>"),
		},
		{
			name:          "data:text/plain;charset=UTF-8;page=21,the%20data:1234,5678",
			rawURL:        "data:text/plain;charset=UTF-8;page=21,the%20data:1234,5678",
			wantMediaType: "text/plain;charset=UTF-8;page=21",
			wantData:      []byte("the data:1234,5678"),
		},
		{
			name:          "data:text/vnd-example+xyz;foo=bar;base64,R0lGODdh",
			rawURL:        "data:text/vnd-example+xyz;foo=bar;base64,R0lGODdh",
			wantMediaType: "text/vnd-example+xyz;foo=bar",
			wantIsBase64:  true,
			wantData:      []byte("GIF87a"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMediaType, gotIsBase64, gotData, err := ParseDataUrls(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDataUrls() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMediaType != tt.wantMediaType {
				t.Errorf("ParseDataUrls() gotMediaType = %v, want %v", gotMediaType, tt.wantMediaType)
			}
			if gotIsBase64 != tt.wantIsBase64 {
				t.Errorf("ParseDataUrls() gotIsBase64 = %v, want %v", gotIsBase64, tt.wantIsBase64)
			}
			if !reflect.DeepEqual(gotData, tt.wantData) {
				t.Errorf("ParseDataUrls() gotData = %v, want %v", gotData, tt.wantData)
			}
		})
	}
}
