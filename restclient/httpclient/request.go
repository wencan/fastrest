package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/wencan/fastrest/restcodecs/restmime"
	"github.com/wencan/fastrest/restcodecs/restvalues"
)

// UrlAddQuery 给url添加查询参数。返回新的url。
// query可以为nil、url.Values，带schema标签的结构体。
// 会覆盖url中原有查询字符串。
func UrlAddQuery(uri string, query interface{}) (string, error) {
	if query == nil {
		return uri, nil
	}

	u, err := url.Parse(uri)
	if err != nil {
		return uri, err
	}

	q, err := restvalues.Encode(query)
	if err != nil {
		return "", err
	}

	u.RawQuery = q
	return u.String(), nil
}

// NewRequestFunc 构建*http.Request的函数。默认为：http.NewRequestWithContext。
type NewRequestFunc func(ctx context.Context, method, url string, body io.Reader) (*http.Request, error)

func newRequestWithQuery(ctx context.Context, method, url string, query interface{}, newRequestFunc NewRequestFunc) (*http.Request, error) {
	method = strings.ToUpper(method)
	switch method {
	case http.MethodGet, http.MethodHead:
		url, err := UrlAddQuery(url, query)
		if err != nil {
			return nil, err
		}
		return newRequestFunc(ctx, method, url, nil)
	default:
		return nil, fmt.Errorf("unsupported method \"%s\"", method)
	}
}

func newRequestWithBody(ctx context.Context, method, url, contentType string, bodyObj interface{}, newRequestFunc NewRequestFunc) (*http.Request, error) {
	method = strings.ToUpper(method)
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		var body io.Reader
		if bodyObj != nil {
			switch t := bodyObj.(type) {
			case io.Reader:
				body = t
			case string:
				body = bytes.NewBufferString(t)
			case []byte:
				body = bytes.NewBuffer(t)
			default:
				var buffer bytes.Buffer
				err := restmime.Marshal(bodyObj, contentType, &buffer)
				if err != nil {
					return nil, err
				}
				body = &buffer
			}
		}

		r, err := newRequestFunc(ctx, method, url, body)
		if err != nil {
			return nil, err
		}
		r.Header.Set("Content-Type", contentType)
		return r, nil
	default:
		return nil, fmt.Errorf("unsupported method \"%s\"", method)
	}
}

// NewRequestWithBody 创建一个带Body的http.Request。
func NewRequestWithBody(ctx context.Context, method, url, contentType string, bodyObj interface{}) (*http.Request, error) {
	return newRequestWithBody(ctx, method, url, contentType, bodyObj, http.NewRequestWithContext)
}
