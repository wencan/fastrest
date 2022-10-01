package restmime

import (
	"errors"
	"io"
	"net/url"

	"github.com/wencan/fastrest/restcodecs/restjson"
	"github.com/wencan/fastrest/restcodecs/restvalues"
)

// MarshalerFunc mime序列化函数签名。
type MarshalerFunc func(v interface{}, writer io.Writer) error

var marshalerMap = map[string]MarshalerFunc{}

// DefaultMarshaler 默认的（保底的）序列化函数。
// 默认为nil，无保底，
var DefaultMarshaler MarshalerFunc

func init() {
	RegisterMarshaler("application/json", jsonMarshaler)
	RegisterMarshaler("application/x-www-form-urlencoded", formMarshaler)
}

// RegisterMarshaler 注册mime数据序列化函数。
func RegisterMarshaler(name string, marshaler MarshalerFunc) {
	marshalerMap[name] = marshaler
}

func jsonMarshaler(v interface{}, writer io.Writer) error {
	encoder := restjson.NewEncoder(writer)
	err := encoder.Encode(v)
	if err != nil {
		return err
	}
	return nil
}

func formMarshaler(v interface{}, writer io.Writer) error {
	values := url.Values{}
	err := restvalues.Encoder.Encode(v, values)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(values.Encode()))
	if err != nil {
		return err
	}
	return nil
}

// Marshal 序列化mime数据。
func Marshal(v interface{}, contentType string, writer io.Writer) error {
	name := ContentTypeName(contentType)
	marshaler := marshalerMap[name]
	if marshaler == nil {
		marshaler = DefaultMarshaler
	}
	if marshaler == nil {
		return errors.New("invalid Content-Type: " + contentType)
	}

	return marshaler(v, writer)
}
