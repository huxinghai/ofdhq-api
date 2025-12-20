package factory

import (
	"ofdhq-api/app/core/container"
	"ofdhq-api/app/global/my_errors"
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/http/validator/core/interf"

	"github.com/gin-gonic/gin"
)

// 表单参数验证器工厂（请勿修改）
func Create(key string) func(context *gin.Context) {

	if value := container.CreateContainersFactory().Get(key); value != nil {
		if val, isOk := value.(interf.ValidatorInterface); isOk {
			return val.CheckParams
		}
	}
	variable.ZapLog.Error(my_errors.ErrorsValidatorNotExists + ", 验证器模块：" + key)
	return nil
}
