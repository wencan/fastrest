package restutils

import (
	"context"
	"log"
	"strings"

	"github.com/wencan/fastrest/restcodecs/restjson"
)

// CompareHumanizeString 比较两个字符串，结果接近可视结果。
func CompareHumanizeString(a, b string) int {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)

	return strings.Compare(a, b)
}

// JsonString 序列化为json，并转为字符串返回。如果序列化失败，返回空字符串。
func JsonString(v interface{}) string {
	data, err := restjson.Marshal(v)
	if err != nil {
		log.Fatalln(err)
	}
	return string(data)
}

// StringFromURL 从url获得字符串。支持的schema：file://、http://、https://。
// url示例：file:///etc/fstab、。
func StringFromURL(ctx context.Context, rawUrl string) (string, error) {
	bytes, err := BytesFromURL(ctx, rawUrl)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
