package restmime

import (
	"errors"
	"io"
	"net/url"

	"github.com/wencan/fastrest/restcodecs/restjson"
	"github.com/wencan/fastrest/restcodecs/restvalues"
)

// UnmarshalerFunc mime反序列化函数签名。
type UnmarshalerFunc func(dest interface{}, reader io.Reader) error

var mimeUnmarshalerMap = map[string]UnmarshalerFunc{}

// DefaultMimeUnmarshaler 默认的（保底的）反序列化函数。
var DefaultMimeUnmarshaler UnmarshalerFunc

func init() {
	RegisterMimeUnmarshaler("application/json", jsonUnmarshaler)
	RegisterMimeUnmarshaler("application/x-www-form-urlencoded", formUnmarshaler)
}

// RegisterMimeUnmarshaler 注册Mime数据反序列化函数。
func RegisterMimeUnmarshaler(name string, unmarshaler UnmarshalerFunc) {
	mimeUnmarshalerMap[name] = unmarshaler
}

func jsonUnmarshaler(dest interface{}, reader io.Reader) error {
	decoder := restjson.NewDecoder(reader)
	err := decoder.Decode(dest)
	if err != nil {
		return err
	}
	return nil
}

func formUnmarshaler(dest interface{}, reader io.Reader) error {
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
	return nil
}

// Unmarshal 反序列化mime数据。
func Unmarshal(dest interface{}, contentType string, reader io.Reader) error {
	name := ContentTypeName(contentType)
	unmarshaler := mimeUnmarshalerMap[name]
	if unmarshaler == nil {
		unmarshaler = DefaultMimeUnmarshaler
	}
	if unmarshaler == nil {
		return errors.New("invalid Content-Type: " + contentType)
	}

	return unmarshaler(dest, reader)
}
