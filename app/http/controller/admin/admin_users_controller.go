package admin

import (
	"errors"
	"fmt"
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/model"
	adminusers "ofdhq-api/app/service/admin_users"
	admintoken "ofdhq-api/app/service/admin_users/token"
	"ofdhq-api/app/utils/cur_userinfo"
	"ofdhq-api/app/utils/md5_encrypt"
	"ofdhq-api/app/utils/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AdminUsers struct {
}

func (u *AdminUsers) Register(context *gin.Context) {
	realName := context.GetString(consts.ValidatorPrefix + "real_name")
	pass := context.GetString(consts.ValidatorPrefix + "pass")
	email := context.GetString(consts.ValidatorPrefix + "email")
	role := int64(context.GetFloat64(consts.ValidatorPrefix + "role_type"))

	olduser, err := model.CreateAdminUserFactory().GetByEmail(email)
	if err != nil {
		response.Fail(context, consts.ServerOccurredErrorCode, consts.ServerOccurredErrorMsg, "")
		return
	}

	if olduser != nil {
		response.Fail(context, consts.ServerOccurredErrorCode, "邮箱地址已经存在使用！", "")
		return
	}
	userIp := context.ClientIP()
	if adminusers.CreateAdminUserFactory().Register(realName, email, pass, userIp, role) {
		response.Success(context, consts.CurdStatusOkMsg, "")
	} else {
		response.Fail(context, consts.CurdRegisterFailCode, consts.CurdRegisterFailMsg, "")
	}
}

func (u *AdminUsers) Login(context *gin.Context) {
	email := context.GetString(consts.ValidatorPrefix + "email")
	pass := context.GetString(consts.ValidatorPrefix + "pass")
	userModelFact := model.CreateAdminUserFactory()
	userModel := userModelFact.Login(email, pass)

	if userModel != nil {
		if userModel.Status <= 0 {
			response.Fail(context, consts.CurdLoginFailCode, "用户已禁封，暂不能登录!", "")
			return
		}

		userTokenFactory := admintoken.CreateUserFactory()
		if userToken, err := userTokenFactory.GenerateToken(userModel.Id, userModel.Email, 432000); err == nil {
			if userTokenFactory.RecordLoginToken(userToken, context.ClientIP()) {

				data := gin.H{
					"admin_user_id": userModel.Id,
					"real_name":     userModel.RealName,
					"email":         userModel.Email,
					"token":         userToken,
					"role_type":     userModel.RoleType,
				}

				response.Success(context, consts.CurdStatusOkMsg, data)
				go userModel.UpdateUserloginInfo(context.ClientIP(), userModel.Id)
				return
			}
		}
	}
	response.Fail(context, consts.CurdLoginFailCode, "用户名与密码错误！", "")
}

func (u *AdminUsers) Update(context *gin.Context) {
	id := int64(context.GetFloat64(consts.ValidatorPrefix + "id"))
	realName := context.GetString(consts.ValidatorPrefix + "real_name")
	email := context.GetString(consts.ValidatorPrefix + "email")
	roleType := int64(context.GetFloat64(consts.ValidatorPrefix + "role_type"))
	status := int64(context.GetFloat64(consts.ValidatorPrefix + "status"))

	adminUser, err := model.CreateAdminUserFactory().GetByAdminUserID(id)
	if err != nil {
		response.Fail(context, consts.ServerOccurredErrorCode, "查询用户失败！", "")
		return
	}

	if adminUser == nil {
		response.Fail(context, consts.ServerOccurredErrorCode, "用户ID找不信息！", "")
		return
	}
	adminUser.Status = status
	adminUser.RealName = realName
	adminUser.Email = email
	adminUser.RoleType = roleType
	err = model.NewAdminUserFactory(adminUser).Update()
	if err != nil {
		variable.ZapLog.Error("更新数据失败！", zap.Error(err))
		response.Fail(context, consts.ServerOccurredErrorCode, consts.ServerOccurredErrorMsg, "")
		return
	}
	response.Success(context, consts.CurdStatusOkMsg, adminUser)
}

func (u *AdminUsers) UpdatePassword(context *gin.Context) {
	id := int64(context.GetFloat64(consts.ValidatorPrefix + "id"))
	password := context.GetString(consts.ValidatorPrefix + "pass")
	adminUser, err := model.CreateAdminUserFactory().GetByAdminUserID(id)
	if err != nil {
		response.Fail(context, consts.ServerOccurredErrorCode, "查询用户失败！", "")
		return
	}

	if adminUser == nil {
		response.Fail(context, consts.ServerOccurredErrorCode, "用户ID找不信息！", "")
		return
	}

	adminUser.Pass = md5_encrypt.Base64Md5(password)
	err = model.NewAdminUserFactory(adminUser).UpdatePassword()
	if err != nil {
		variable.ZapLog.Error("更新数据失败！", zap.Error(err))
		response.Fail(context, consts.ServerOccurredErrorCode, consts.ServerOccurredErrorMsg, "")
		return
	}
	response.Success(context, consts.CurdStatusOkMsg, adminUser)
}

func (u *AdminUsers) List(ctx *gin.Context) {
	page := int(ctx.GetFloat64(consts.ValidatorPrefix + "page"))
	limit := int(ctx.GetFloat64(consts.ValidatorPrefix + "limit"))

	result, err := model.CreateAdminUserFactory().List(int64(page), int64(limit))
	if err != nil {
		variable.ZapLog.Error("更新数据失败！", zap.Error(err))
		response.Fail(ctx, consts.ServerOccurredErrorCode, consts.ServerOccurredErrorMsg, "")
		return
	}

	pageInfo := response.GenPageInfo(page, limit, len(result))
	response.Success(ctx, consts.CurdStatusOkMsg, gin.H{"page_info": pageInfo, "counts": len(result), "list": result})
}

func (u *AdminUsers) buildCurrentUser(ctx *gin.Context) (*model.AdminUsers, error) {
	adminUserID, exist := cur_userinfo.GetCurrentAdminUserId(ctx)
	if !exist {
		return nil, errors.New("没有当前用户")
	}
	userFactory := model.CreateAdminUserFactory()
	user, err := userFactory.GetByAdminUserID(adminUserID)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("获取Admin用户信息失败！userID:%d", adminUserID))
	}

	if user == nil {
		return nil, errors.Join(err, fmt.Errorf("没有查到Admin用户的数据！userID:%d", adminUserID))
	}

	return user, nil
}
