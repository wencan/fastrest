package restvalues

import (
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/gorilla/schema"
)

// Encoder url.Values编码器。需要结构体字段带schema标签。
var Encoder = schema.NewEncoder()

func init() {
	Encoder.RegisterEncoder(time.Time{}, func(v reflect.Value) string {
		fmt.Println("time.Time{}")
		return v.Interface().(time.Time).Format(time.RFC3339)
	})
	Encoder.RegisterEncoder(&time.Time{}, func(v reflect.Value) string {
		fmt.Println("&time.Time{}")
		if v.IsNil() {
			return ""
		}
		return v.Interface().(*time.Time).Format(time.RFC3339)
	})
}

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
