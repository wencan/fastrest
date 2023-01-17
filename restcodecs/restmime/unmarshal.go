package restmime

import (
	"errors"
	"fmt"
	"io"
	"mime"

	"github.com/wencan/fastrest/restcodecs/restjson"
	"github.com/wencan/fastrest/restcodecs/restvalues"
	"google.golang.org/protobuf/proto"
)

// UnmarshalerFunc mime反序列化函数签名。
type UnmarshalerFunc func(dest interface{}, reader io.Reader) error

var unmarshalerMap = map[string]UnmarshalerFunc{}

func init() {
	RegisterUnmarshaler(string(MimeTypeJson), JsonUnmarshaler)
	RegisterUnmarshaler(string(MimeTypeForm), FormUnmarshaler)
	RegisterUnmarshaler(string(MimeTypeProtobuf), ProtobufUnmarshaler)
}

// RegisterUnmarshaler 注册Mime数据反序列化函数。
func RegisterUnmarshaler(name string, unmarshaler UnmarshalerFunc) {
	unmarshalerMap[name] = unmarshaler
}

// JsonUnmarshaler 反序列化json。
func JsonUnmarshaler(dest interface{}, reader io.Reader) error {
	decoder := restjson.NewDecoder(reader)
	err := decoder.Decode(dest)
	if err != nil {
		return err
	}
	return nil
}

// FormUnmarshaler 反序列化查询/表单。需要dest字段带schema标签。
func FormUnmarshaler(dest interface{}, reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	return restvalues.Decode(dest, string(data))
}

// ProtobufUnmarshaler 反序列化Protocol Buffers消息。
func ProtobufUnmarshaler(dest interface{}, reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	message, ok := dest.(proto.Message)
	if !ok {
		return errors.New("not protobuf message")
	}
	err = proto.Unmarshal(data, message)
	if err != nil {
		return err
	}
	return nil
}

// Unmarshal 反序列化mime数据。
func Unmarshal(dest interface{}, contentType string, reader io.Reader) error {
	name, _, _ := mime.ParseMediaType(contentType)
	if name == "" {
		return fmt.Errorf("wrong content type: [%s]", contentType)
	}
	unmarshaler := unmarshalerMap[name]
	if unmarshaler == nil {
		return fmt.Errorf("invalid content type: [%s]", contentType)
	}

	return unmarshaler(dest, reader)
}
