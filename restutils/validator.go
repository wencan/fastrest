package restutils

import (
	"context"
	"reflect"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateStruct 校验结构体对象字段值。
// 依赖于字段validate标签。校验支持：https://github.com/go-playground/validator。
func ValidateStruct(ctx context.Context, s interface{}) error {
	value := reflect.ValueOf(s)
	switch value.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			elem := value.Index(i)
			err := ValidateStruct(ctx, elem.Interface())
			if err != nil {
				return err
			}
		}
	default:
		return validate.StructCtx(ctx, s)
	}

	return nil
}
