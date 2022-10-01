package restjson

import (
	jsoniter "github.com/json-iterator/go"
)

// Marshal json序列化函数。可以覆盖。
var Marshal = jsoniter.ConfigCompatibleWithStandardLibrary.Marshal

// Marshal json反序列化函数。可以覆盖。
var Unmarshal = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal

// NewEncoder 创建json序列化编码器。可以覆盖。
var NewEncoder = jsoniter.ConfigCompatibleWithStandardLibrary.NewEncoder

// NewDecoder 创建json序列化解码器。可以覆盖。
var NewDecoder = jsoniter.ConfigCompatibleWithStandardLibrary.NewDecoder
