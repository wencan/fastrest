package restutils

import (
	"context"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateStruct 校验结构体对象字段值。
// 依赖于字段validate标签。校验支持：https://github.com/go-playground/validator。
func ValidateStruct(ctx context.Context, s interface{}) error {
	return validate.StructCtx(ctx, s)
}
