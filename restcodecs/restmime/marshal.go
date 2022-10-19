package restmime

import (
	"errors"
	"io"
	"net/url"

	"github.com/wencan/fastrest/restcodecs/restjson"
	"github.com/wencan/fastrest/restcodecs/restvalues"
	"google.golang.org/protobuf/proto"
)

// MarshalerFunc mime序列化函数签名。
type MarshalerFunc func(v interface{}, writer io.Writer) error

var marshalerMap = map[string]MarshalerFunc{}

// DefaultMarshaler 默认的（保底的）序列化函数。可覆盖。
// 默认保底序列化为json，
var DefaultMarshaler MarshalerFunc = jsonMarshaler

func init() {
	RegisterMarshaler(string(MimeTypeJson), jsonMarshaler)
	RegisterMarshaler(string(MimeTypeForm), formMarshaler)
	RegisterMarshaler(string(MimeTypeProtobuf), protobufMarshler)
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

func protobufMarshler(v interface{}, writer io.Writer) error {
	message, ok := v.(proto.Message)
	if !ok {
		return errors.New("not protobuf message")
	}
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
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