package routers

import (
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/http/middleware/authorization"
	validatorFactory "ofdhq-api/app/http/validator/core/factory"

	"github.com/gin-gonic/gin"
)

func InitAdminApiRouter(router *gin.Engine) {
	vApi := router.Group("/api/v1/admins/")
	{
		vApi.POST("login", validatorFactory.Create(consts.ValidatorPrefix+"AdminUserLogin"))
		vApi.Use(authorization.CheckAdminTokenAuth())
		vApi.POST("create_user", validatorFactory.Create(consts.ValidatorPrefix+"AdminUserRegister"))
		vApi.POST("update_user", validatorFactory.Create(consts.ValidatorPrefix+"AdminAdminUserUpdate"))
		vApi.POST("update_user_password", validatorFactory.Create(consts.ValidatorPrefix+"AdminUserUpdatePassword"))
		vApi.GET("list_user", validatorFactory.Create(consts.ValidatorPrefix+"AdminAdminUserList"))

		topic := vApi.Group("topic/")
		{
			topic.POST("create", validatorFactory.Create(consts.ValidatorPrefix+"AdminTopicCreate"))
			topic.POST("delete", validatorFactory.Create(consts.ValidatorPrefix+"AdminTopicDelete"))
			topic.POST("update", validatorFactory.Create(consts.ValidatorPrefix+"AdminTopicUpdate"))
			topic.GET("list", validatorFactory.Create(consts.ValidatorPrefix+"AdminTopicList"))
		}
	}
}
