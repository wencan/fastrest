package restutils

import (
	"errors"
	"io"
	"mime"
	"net/url"
	"strings"

	"github.com/wencan/fastrest/restcodecs/restjson"
	"github.com/wencan/fastrest/restcodecs/restvalues"
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
	if len(names) > 1 {
		return names[0]
	}
	return ""
}

// UnmarshalContent 反序列化指定Type的Content。
// 支持application/json和application/x-www-form-urlencoded。
func UnmarshalContent(dest interface{}, contentType string, reader io.Reader) error {
	typeName := ContentTypeName(contentType)
	switch typeName {
	case "application/json":
		decoder := restjson.NewDecoder(reader)
		err := decoder.Decode(dest)
		if err != nil {
			return err
		}
	case "application/x-www-form-urlencoded":
		data, err := io.ReadAll(reader)
		if err != nil {
			return err
		}
		values, err := url.ParseQuery(string(data))
		if err != nil {
			return err
		}
		err = restvalues.Decoder.Decode(dest, values)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid Content-Type: " + contentType)
	}

	return nil
}
