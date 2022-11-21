package restmime

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/url"

	"github.com/wencan/fastrest/restcodecs/restjson"
	"github.com/wencan/fastrest/restcodecs/restvalues"
	"google.golang.org/protobuf/proto"
)

// UnmarshalerFunc mime反序列化函数签名。
type UnmarshalerFunc func(dest interface{}, reader io.Reader) error

var unmarshalerMap = map[string]UnmarshalerFunc{}

func init() {
	RegisterUnmarshaler(string(MimeTypeJson), jsonUnmarshaler)
	RegisterUnmarshaler(string(MimeTypeForm), formUnmarshaler)
	RegisterUnmarshaler(string(MimeTypeProtobuf), protobufUnmarshaler)
}

// RegisterUnmarshaler 注册Mime数据反序列化函数。
func RegisterUnmarshaler(name string, unmarshaler UnmarshalerFunc) {
	unmarshalerMap[name] = unmarshaler
}

func jsonUnmarshaler(dest interface{}, reader io.Reader) error {
	decoder := restjson.NewDecoder(reader)
	err := decoder.Decode(dest)
	if err != nil {
		return err
	}
	return nil
}

// formUnmarshaler 使用github.com/gorilla/schema反序列化数据。需要dest字段带schema标签。
func formUnmarshaler(dest interface{}, reader io.Reader) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	values, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	destValues, _ := dest.(*url.Values)
	if destValues != nil {
		*destValues = values
	} else {
		err = restvalues.Decoder.Decode(dest, values)
		if err != nil {
			return err
		}
	}
	return nil
}

func protobufUnmarshaler(dest interface{}, reader io.Reader) error {
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
