package resterror

import (
	"github.com/wencan/fastrest/restutils"
)

// // ThirdpartyUnwrap 解除第三方包装的方法。
// var ThirdpartyUnwrap func(error) error

// // Unwrap 解包装
// func Unwrap(err error) error {
// 	if ThirdpartyUnwrap != nil {
// 		err = ThirdpartyUnwrap(err)
// 	}
// 	return err
// }

// FixNilError 如果err指向是一个空指针，返回nil。
func FixNilError(err error) error {
	if err == nil {
		return nil
	}
	if restutils.IsGhostInterface(err) {
		return nil
	}
	return err
}
