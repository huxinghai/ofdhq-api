package authorization

import (
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/global/variable"

	admintoken "ofdhq-api/app/service/admin_users/token"
	"ofdhq-api/app/utils/response"
	"strings"

	"github.com/gin-gonic/gin"
)

// CheckTokenAuth 检查token完整性、有效性中间件
func CheckAdminTokenAuth() gin.HandlerFunc {
	return func(context *gin.Context) {

		headerParams := HeaderParams{}

		//  推荐使用 ShouldBindHeader 方式获取头参数
		if err := context.ShouldBindHeader(&headerParams); err != nil {
			response.TokenErrorParam(context, consts.JwtTokenMustValid+err.Error())
			return
		}
		token := strings.Split(headerParams.Authorization, " ")
		if len(token) == 2 && len(token[1]) >= 20 {
			tokenIsEffective := admintoken.CreateUserFactory().IsEffective(token[1])
			if tokenIsEffective {
				if customToken, err := admintoken.CreateUserFactory().ParseToken(token[1]); err == nil {
					key := variable.AdminBindContextKeyName
					// token验证通过，同时绑定在请求上下文
					context.Set(key, customToken)
				}
				context.Next()
			} else {
				response.ErrorTokenAuthFail(context)
			}
		} else {
			response.ErrorTokenBaseInfo(context)
		}
	}
}
