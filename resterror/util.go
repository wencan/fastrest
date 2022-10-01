package resterror

import (
	"github.com/wencan/fastrest/restutils"
)

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
