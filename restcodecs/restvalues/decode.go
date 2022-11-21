package restvalues

import (
	"net/url"

	"github.com/gorilla/schema"
)

// Decoder url.Values解码器。需要结构体字段带schema标签。
var Decoder = schema.NewDecoder()

func init() {
	Decoder.IgnoreUnknownKeys(true)
}

// Decode 解码表单/查询字符串。支持*url.Values和带schema标签的结构体指针。
func Decode(dest interface{}, str string) error {
	values, err := url.ParseQuery(str)
	if err != nil {
		return err
	}

	destValues, _ := dest.(*url.Values)
	if destValues != nil {
		*destValues = values
	} else {
		err = Decoder.Decode(dest, values)
		if err != nil {
			return err
		}
	}

	return nil
}
