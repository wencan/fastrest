package restvalues

import (
	"net/url"

	"github.com/gorilla/schema"
)

// Encoder url.Values编码器。需要结构体字段带schema标签。
var Encoder = schema.NewEncoder()

// Encode 编码表单/查询字符串，支持url.Values和带schema的结构体和结构体指针。
func Encode(v interface{}) (string, error) {
	values, _ := v.(url.Values)
	if values == nil {
		values = url.Values{}
		err := Encoder.Encode(v, values)
		if err != nil {
			return "", err
		}
	}
	return values.Encode(), nil
}
