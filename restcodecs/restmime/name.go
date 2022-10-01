package restmime

import (
	"mime"
	"strings"
)

// ContentTypeNames 返回Content Type的名称列表。不含参数。
// 有效的Content Type格式如下：
// <MIME_type>/<MIME_subtype>
// <MIME_type>/*
// */*
// text/html, application/xhtml+xml, application/xml;q=0.9, image/webp, */*;q=0.8
// 如果没有有效的Content Type，返回空字符串。
func ContentTypeNames(contentType string) []string {
	var names []string
	parts := strings.Split(contentType, ",") // 多个类型用,分割
	for _, part := range parts {
		part = strings.TrimSpace(part)
		mimeType, _, _ := mime.ParseMediaType(part)
		if mimeType != "" {
			names = append(names, mimeType)
		}
	}

	return names
}

// ContentTypeName 返回Content Type的名称。如果有多个Content Type名称，只返回第一个。
func ContentTypeName(contentType string) string {
	names := ContentTypeNames(contentType)
	if len(names) >= 1 {
		return names[0]
	}
	return ""
}
