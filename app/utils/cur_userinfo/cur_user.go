package cur_userinfo

import (
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/http/middleware/my_jwt"

	"github.com/gin-gonic/gin"
)

// GetCurrentUserId 获取当前用户的id
// @context 请求上下文
func GetCurrentUserId(context *gin.Context) (int64, bool) {
	tokenKey := variable.ConfigYml.GetString("Token.BindContextKeyName")
	currentUser, exist := context.MustGet(tokenKey).(my_jwt.CustomClaims)
	return currentUser.UserId, exist
}

func IsExistCurrentUserID(context *gin.Context) bool {
	tokenKey := variable.ConfigYml.GetString("Token.BindContextKeyName")
	_, exist := context.Get(tokenKey)
	return exist
}

// GetCurrentUserId 获取当前用户的id
// @context 请求上下文
func GetCurrentAdminUserId(context *gin.Context) (int64, bool) {
	tokenKey := variable.AdminBindContextKeyName
	currentUser, exist := context.MustGet(tokenKey).(my_jwt.AdminCustomClaims)
	return currentUser.AdminUserId, exist
}
