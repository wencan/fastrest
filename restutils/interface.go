package restutils

import "reflect"

// IsGhostInterface 判断接口是否为nil，或者指向指针是否为空。
// 使用指向空指针的接口去调函数，会panic。
func IsGhostInterface(i interface{}) bool {
	if i == nil {
		return true
	}

	value := reflect.ValueOf(i)
	switch value.Kind() {
	case reflect.Ptr:
		target := reflect.Indirect(value)
		return !target.IsValid()
	default:
	}

	return false
}
