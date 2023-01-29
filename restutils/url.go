package restutils

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

// ParseDataUrls 解析Data Urls。
// 协议参考：https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/Data_URLs 、 https://en.wikipedia.org/wiki/Data_URI_scheme。
// 格式：data:[<mediatype>][;base64],<data>
func ParseDataUrls(rawURL string) (mediaType string, isBase64 bool, data []byte, err error) {
	if strings.Index(rawURL, "data:") != 0 {
		return "", false, nil, fmt.Errorf("invalid scheme")
	}

	// 先去掉data:
	path := rawURL[len("data:"):]

	// 分割mediatype+base64和数据
	argsAndData := strings.Split(path, ",")
	if len(argsAndData) < 2 {
		return "", false, nil, fmt.Errorf("invalid data url")
	}
	encodedData := path[len(argsAndData[0])+1:]

	// mediatype和base64部分
	// mediatype包含参数
	if strings.HasSuffix(argsAndData[0], ";base64") {
		isBase64 = true
		mediaType = argsAndData[0][:len(argsAndData[0])-7] // 去掉后面的;base64
	} else {
		mediaType = argsAndData[0]
	}

	if isBase64 {
		data, err = base64.StdEncoding.DecodeString(encodedData)
		if err != nil {
			return "", false, nil, fmt.Errorf("failed in decode base64 data")
		}
	} else {
		d, err := url.PathUnescape(encodedData)
		if err != nil {
			return "", false, nil, fmt.Errorf("failed in unescape path data")
		}
		data = []byte(d)
	}

	return
}
