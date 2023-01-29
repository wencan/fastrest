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

// StringFromURL 从url获得字符串。
// 支持的scheme：file://、http://、https://、data:。
// url示例：file:///etc/fstab、data:text/plain;base64,SGVsbG8sIFdvcmxkIQ==。
func StringFromURL(ctx context.Context, rawUrl string) (string, error) {
	bytes, err := BytesFromURL(ctx, rawUrl)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// SubstringsWithTags 根据前后标签，找出子字符串。
// 比如字符串为“我是一名{[(程序员)]}”，前后标签为“{[(”和“)]}”，匹配的字符串为“程序员”。
// 返回元素顺序同它们在字符串内的顺序。
// 如果存在重复的匹配子字符串，函数返回元素也将重复。
func SubstringsWithTags(str, startTag, endTag string) []string {
	var substrs []string
	parts := strings.Split(str, startTag)
	for idx, part := range parts {
		if idx == 0 { // 第一个应该是空字符串
			continue
		}
		if part == "" {
			continue
		}
		// if !strings.Contains(part, endTag) {
		// 	continue
		// }
		subs := strings.Split(part, endTag)
		if len(subs) <= 1 {
			continue
		}
		substrs = append(substrs, subs[0])
	}
	return substrs
}
