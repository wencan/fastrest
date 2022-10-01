package restvalues

import (
	"github.com/gorilla/schema"
)

// Encoder url.Values编码器。
var Encoder = schema.NewEncoder()

// Decoder url.Values解码器。
var Decoder = schema.NewDecoder()

func init() {
	Decoder.IgnoreUnknownKeys(true)
}
