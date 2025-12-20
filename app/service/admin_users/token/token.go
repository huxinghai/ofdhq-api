package token

import (
	"errors"
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/global/my_errors"
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/http/middleware/my_jwt"
	"ofdhq-api/app/model"

	"github.com/dgrijalva/jwt-go"
)

// CreateUserFactory 创建 userToken 工厂
func CreateUserFactory() *adminUserToken {
	return &adminUserToken{
		adminJwt: my_jwt.CreateAdminWT("ofdhq-api-admin"),
	}
}

type adminUserToken struct {
	adminJwt *my_jwt.AdminJwtSign
}

// GenerateToken 生成token
func (u *adminUserToken) GenerateToken(adminUserid int64, email string, expireAt int64) (tokens string, err error) {

	// 根据实际业务自定义token需要包含的参数，生成token，注意：用户密码请勿包含在token
	customClaims := my_jwt.AdminCustomClaims{
		AdminUserId: adminUserid,
		Email:       email,
		// 特别注意，针对前文的匿名结构体，初始化的时候必须指定键名，并且不带 jwt. 否则报错：Mixture of field: value and value initializers
		StandardClaims: jwt.StandardClaims{
			NotBefore: variable.NowTimeSH().Unix() - 10,       // 生效开始时间
			ExpiresAt: variable.NowTimeSH().Unix() + expireAt, // 失效截止时间
		},
	}
	return u.adminJwt.CreateToken(customClaims)
}

// RecordLoginToken 用户login成功，记录用户token
func (u *adminUserToken) RecordLoginToken(userToken, clientIp string) bool {
	if customClaims, err := u.adminJwt.ParseToken(userToken); err == nil {
		adminUserId := customClaims.AdminUserId
		expiresAt := customClaims.ExpiresAt
		return model.CreateAdminUserFactory().OauthLoginToken(adminUserId, userToken, expiresAt, clientIp)
	} else {
		return false
	}
}

// TokenIsMeetRefreshCondition 检查token是否满足刷新条件
func (u *adminUserToken) TokenIsMeetRefreshCondition(token string) bool {
	// token基本信息是否有效：1.过期时间在允许的过期范围内;2.基本格式正确
	customClaims, code := u.isNotExpired(token, variable.ConfigYml.GetInt64("Token.JwtTokenRefreshAllowSec"))
	switch code {
	case consts.JwtTokenOK, consts.JwtTokenExpired:
		//在数据库的存储信息是否也符合过期刷新刷新条件
		if model.CreateAdminUserFactory().OauthRefreshConditionCheck(customClaims.AdminUserId, token) {
			return true
		}
	}
	return false
}

// RefreshToken 刷新token的有效期（默认+3600秒，参见常量配置项）
func (u *adminUserToken) RefreshToken(oldToken, clientIp string) (newToken string, res bool) {
	var err error
	//如果token是有效的、或者在过期时间内，那么执行更新，换取新token
	if newToken, err = u.adminJwt.RefreshToken(oldToken, 432000); err == nil {
		if customClaims, err := u.adminJwt.ParseToken(newToken); err == nil {
			adminUserId := customClaims.AdminUserId
			expiresAt := customClaims.ExpiresAt
			if model.CreateAdminUserFactory().OauthRefreshToken(adminUserId, expiresAt, oldToken, newToken, clientIp) {
				return newToken, true
			}
		}
	}

	return "", false
}

// 判断token本身是否未过期
// 参数解释：
// token： 待处理的token值
// expireAtSec： 过期时间延长的秒数，主要用于用户刷新token时，判断是否在延长的时间范围内，非刷新逻辑默认为0
func (u *adminUserToken) isNotExpired(token string, expireAtSec int64) (*my_jwt.AdminCustomClaims, int) {
	if customClaims, err := u.adminJwt.ParseToken(token); err == nil {

		if variable.NowTimeSH().Unix()-(customClaims.ExpiresAt+expireAtSec) < 0 {
			// token有效
			return customClaims, consts.JwtTokenOK
		} else {
			// 过期的token
			return customClaims, consts.JwtTokenExpired
		}
	} else {
		// 无效的token
		return nil, consts.JwtTokenInvalid
	}
}

// IsEffective 判断token是否有效（未过期+数据库用户信息正常）
func (u *adminUserToken) IsEffective(token string) bool {
	customClaims, code := u.isNotExpired(token, 0)
	if consts.JwtTokenOK == code {
		if model.CreateAdminUserFactory().OauthCheckTokenIsOk(customClaims.AdminUserId, token) {
			return true
		}
	}
	return false
}

// ParseToken 将 token 解析为绑定时传递的参数
func (u *adminUserToken) ParseToken(tokenStr string) (CustomClaims my_jwt.AdminCustomClaims, err error) {
	if customClaims, err := u.adminJwt.ParseToken(tokenStr); err == nil {
		return *customClaims, nil
	} else {
		return my_jwt.AdminCustomClaims{}, errors.New(my_errors.ErrorsParseTokenFail)
	}
}

// DestroyToken 销毁token，基本用不到，因为一个网站的用户退出都是直接关闭浏览器窗口，极少有户会点击“注销、退出”等按钮，销毁token其实无多大意义
func (u *adminUserToken) DestroyToken() {

}
