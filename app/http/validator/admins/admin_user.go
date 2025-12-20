package admins

import (
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/http/controller/admin"
	common_data_type "ofdhq-api/app/http/validator/common/data_type"
	"ofdhq-api/app/http/validator/core/data_transfer"
	"ofdhq-api/app/utils/common"
	"ofdhq-api/app/utils/response"

	"github.com/gin-gonic/gin"
)

type UserLogin struct {
	Email string `form:"email" json:"email"  binding:"required,min=2"`
	Pass  string `form:"pass" json:"pass" binding:"required,min=6,max=20"`
}

func (l UserLogin) CheckParams(context *gin.Context) {
	if err := context.ShouldBind(&l); err != nil {
		response.ValidatorError(context, err)
		return
	}
	extraAddBindDataContext := data_transfer.DataAddContext(l, consts.ValidatorPrefix, context)
	if extraAddBindDataContext == nil {
		response.ErrorSystem(context, "userLogin表单验证器json化失败", "")
	} else {
		(&admin.AdminUsers{}).Login(extraAddBindDataContext)
	}
}

type UserRegister struct {
	RealName string `form:"real_name" json:"real_name" binding:"required,min=1"`
	Email    string `form:"email" json:"email"  binding:"required,min=2"`
	Pass     string `form:"pass" json:"pass" binding:"required,min=6,max=20"`
	RoleType int64  `form:"role_type" json:"role_type"`
}

func (l UserRegister) CheckParams(context *gin.Context) {
	if err := context.ShouldBind(&l); err != nil {
		response.ValidatorError(context, err)
		return
	}
	extraAddBindDataContext := data_transfer.DataAddContext(l, consts.ValidatorPrefix, context)
	if extraAddBindDataContext == nil {
		response.ErrorSystem(context, "userLogin表单验证器json化失败", "")
	} else {
		(&admin.AdminUsers{}).Register(extraAddBindDataContext)
	}
}

type UserUpdate struct {
	ID       int64  `form:"id" json:"id"`
	RealName string `form:"real_name" json:"real_name" binding:"required,min=1"`
	Email    string `form:"email" json:"email"  binding:"required,min=2"`
	RoleType int64  `form:"role_type" json:"role_type"`
	Status   int64  `form:"status" json:"status"`
}

func (l UserUpdate) CheckParams(context *gin.Context) {
	if err := context.ShouldBind(&l); err != nil {
		response.ValidatorError(context, err)
		return
	}
	if !common.ContainsInt(l.Status, []int64{0, 1}) {
		response.ErrorParam(context, "状态传入有误", nil)
		return
	}
	extraAddBindDataContext := data_transfer.DataAddContext(l, consts.ValidatorPrefix, context)
	if extraAddBindDataContext == nil {
		response.ErrorSystem(context, "UserUpdate表单验证器json化失败", "")
	} else {
		(&admin.AdminUsers{}).Update(extraAddBindDataContext)
	}
}

type UserUpdatePassword struct {
	ID          int64  `form:"id" json:"id"`
	Pass        string `form:"pass" json:"pass" binding:"required,min=6,max=20"`
	ConfirmPass string `form:"confirm_pass" json:"confirm_pass" binding:"required,min=6,max=20"`
}

func (l UserUpdatePassword) CheckParams(context *gin.Context) {
	if err := context.ShouldBind(&l); err != nil {
		response.ValidatorError(context, err)
		return
	}

	if l.Pass != l.ConfirmPass {
		response.ErrorParam(context, "两次输入密码不一致！", nil)
		return
	}

	extraAddBindDataContext := data_transfer.DataAddContext(l, consts.ValidatorPrefix, context)
	if extraAddBindDataContext == nil {
		response.ErrorSystem(context, "UserUpdate表单验证器json化失败", "")
	} else {
		(&admin.AdminUsers{}).UpdatePassword(extraAddBindDataContext)
	}
}

type AdminUserList struct {
	common_data_type.Page
}

func (l AdminUserList) CheckParams(context *gin.Context) {
	l.Page.SetDefault()
	if err := context.ShouldBind(&l); err != nil {
		response.ValidatorError(context, err)
		return
	}

	extraAddBindDataContext := data_transfer.DataAddContext(l, consts.ValidatorPrefix, context)
	if extraAddBindDataContext == nil {
		response.ErrorSystem(context, "UserUpdate表单验证器json化失败", "")
	} else {
		(&admin.AdminUsers{}).List(extraAddBindDataContext)
	}
}
