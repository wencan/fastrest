package restutils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// BytesFromURL 从url获得数据。
// 支持的scheme：file://、http://、https://、data:。
// url示例：file:///etc/fstab、data:text/plain;base64,SGVsbG8sIFdvcmxkIQ==。
func BytesFromURL(ctx context.Context, rawUrl string) ([]byte, error) {
	file, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	switch file.Scheme {
	case "file":
		f, err := os.Open(file.Path)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		data, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}
		return data, nil
	case "http", "https":
		r, err := http.NewRequestWithContext(ctx, http.MethodGet, rawUrl, nil)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("cannot get url, status: %d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return data, nil
	case "data":
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/Data_URLs
		// https://en.wikipedia.org/wiki/Data_URI_scheme
		// 格式：data:[<mediatype>][;base64],<data>
		_, _, data, err := ParseDataUrls(rawUrl)
		if err != nil {
			return nil, err
		}
		return data, nil
	default:
		return nil, fmt.Errorf("not supported scheme: [%s]", file.Scheme)
	}
}
