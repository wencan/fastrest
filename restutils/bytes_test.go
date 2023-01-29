package restutils

import (
	"context"
	"os"
	"reflect"
	"testing"
)

func TestBytesFromURL(t *testing.T) {
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
			name:   "https_page",
			rawUrl: "https://www.w3.org/MarkUp/html-test/implementors-guide/plaintext.html",
			want: []byte(`<title>plaintext test</title>

<H1>Is plaintext supported in any way?</H1>

What happens after this?

On Mosaic, everything after the <code>&lt;PLAINTEXT&gt;</code>
tag is treated like a text/plain body part -- not SGML at
all. Nothing is recognized: not even a <code>PLAINTEXT</code>
end tag.<p>

On linemode WWW, it works just like XMP -- i.e. it's like CDATA,
except that <code>&lt;/</code> is only recognized when follwed
by <code>PLAINTEXT</code>.

<PLAINTEXT>

lkjsdflkj

We need the following so it'll parse...

</plaintext>
`)},
		{
			name:   "data_url",
			rawUrl: "data:,Hello%2C%20World%21",
			want:   []byte("Hello, World!"),
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
