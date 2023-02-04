package restutils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestBytesFromURL(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("测试Body数据"))
	}))
	defer s.Close()

	tests := []struct {
		name    string
		rawUrl  string
		want    []byte
		wantErr bool
	}{
		{
			name: "local_file",
			rawUrl: func() string {
				file, err := os.CreateTemp("", "unittest")
				if err != nil {
					panic(err)
				}
				defer file.Close()
				_, err = file.Write([]byte("test 测试"))
				if err != nil {
					panic(err)
				}
				return "file://" + file.Name()
			}(),
			want: []byte("test 测试"),
		},
		{
			name:   "http_page",
			rawUrl: s.URL,
			want:   []byte(`测试Body数据`),
		},
		{
			name:   "data_url",
			rawUrl: "data:,Hello%2C%20World%21",
			want:   []byte("Hello, World!"),
		},
		{
			name:   "data_url_with_parameters",
			rawUrl: "data:text/plain;charset=UTF-8;page=21,the%20data:1234,5678",
			want:   []byte("the data:1234,5678"),
		},
		{
			name:   "base64_data_url",
			rawUrl: "data:text/plain;base64,SGVsbG8sIFdvcmxkIQ==",
			want:   []byte("Hello, World!"),
		},
		{
			name:    "invalid_scheme",
			rawUrl:  "test://123",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BytesFromURL(context.TODO(), tt.rawUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("BytesFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BytesFromURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
