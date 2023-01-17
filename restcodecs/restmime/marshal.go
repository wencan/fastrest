package restmime

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"strings"

	"github.com/wencan/fastrest/restcodecs/restjson"
	"github.com/wencan/fastrest/restcodecs/restvalues"
	"google.golang.org/protobuf/proto"
)

// MarshalerFunc mime序列化函数签名。
type MarshalerFunc func(v interface{}, writer io.Writer) error

type registeredMarshaler struct {
	ContentType  string
	DiscreteType string
	Marshaler    MarshalerFunc
}

var registeredMarshalers = make([]*registeredMarshaler, 0)

func init() {
	RegisterMarshaler(string(MimeTypeJson), JsonMarshaler)
	RegisterMarshaler(string(MimeTypeForm), FormMarshaler)
	RegisterMarshaler(string(MimeTypeProtobuf), ProtobufMarshler)
}

// RegisterMarshaler 注册mime数据序列化函数。
func RegisterMarshaler(name string, marshaler MarshalerFunc) {
	registeredMarshalers = append(registeredMarshalers, &registeredMarshaler{
		ContentType:  name,
		DiscreteType: strings.Split(name, "/")[0],
		Marshaler:    marshaler,
	})
}

// JsonMarshaler 序列化json。
func JsonMarshaler(v interface{}, writer io.Writer) error {
	encoder := restjson.NewEncoder(writer)
	err := encoder.Encode(v)
	if err != nil {
		return err
	}
	return nil
}

// FormMarshaler 序列化查询/表单。v为url.Values或者结构体字段带schema标签。
func FormMarshaler(v interface{}, writer io.Writer) error {
	str, err := restvalues.Encode(v)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(str))
	if err != nil {
		return err
	}
	return nil
}

// ProtobufMarshler 序列化Protocol Buffers消息。
func ProtobufMarshler(v interface{}, writer io.Writer) error {
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
	name, _, _ := mime.ParseMediaType(contentType)
	if name == "" {
		return fmt.Errorf("wrong content type: [%s]", contentType)
	}

	var marshaler MarshalerFunc
	for _, registeredMarshaler := range registeredMarshalers {
		if registeredMarshaler.ContentType == name {
			marshaler = registeredMarshaler.Marshaler
			break
		}
	}
	if marshaler == nil {
		return errors.New("invalid content type: [%s]" + contentType)
	}

	return marshaler(v, writer)
}

// AcceptableMarshalContentType 根据accept要求，在支持的mime type列表中寻找匹配的content type。
// accept举例：text/html, application/xhtml+xml, application/xml;q=0.9, image/webp, */*;q=0.8。
// 目前未支持Accept中的参数部分。
func AcceptableMarshalContentType(accept string) string {
	acceptParts := strings.Split(accept, ",")
	for _, acceptPart := range acceptParts {
		acceptPart, _, _ := mime.ParseMediaType(acceptPart)
		if acceptPart == "*/*" {
			return registeredMarshalers[0].ContentType
		} else if strings.Contains(acceptPart, "/*") {
			discreteType := strings.Split(acceptPart, "/")[0]
			for _, registeredMarshaler := range registeredMarshalers {
				if registeredMarshaler.DiscreteType == discreteType {
					return registeredMarshaler.ContentType
				}
			}
		} else {
			for _, registeredMarshaler := range registeredMarshalers {
				if registeredMarshaler.ContentType == acceptPart {
					return registeredMarshaler.ContentType
				}
			}
		}
	}

	return ""
}
