package restvalues

import (
	"github.com/gorilla/schema"
)

// Encoder url.Values编码器。需要结构体字段带schema标签。
var Encoder = schema.NewEncoder()

// Decoder url.Values解码器。需要结构体字段带schema标签。
var Decoder = schema.NewDecoder()

func init() {
	Decoder.IgnoreUnknownKeys(true)
}
