package admins

import (
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/http/controller/admin"
	"ofdhq-api/app/http/controller/api"
	common_data_type "ofdhq-api/app/http/validator/common/data_type"
	"ofdhq-api/app/http/validator/core/data_transfer"
	"ofdhq-api/app/utils/response"

	"github.com/gin-gonic/gin"
)

type AdminTopicCreate struct {
	Title string `form:"title" json:"title" binding:"required,min=1"`
	Body  string `form:"body" json:"body" binding:"required,min=1"`
}

func (l AdminTopicCreate) CheckParams(ctx *gin.Context) {
	if err := ctx.ShouldBind(&l); err != nil {
		response.ValidatorError(ctx, err)
		return
	}

	extraContext := data_transfer.DataAddContext(l, consts.ValidatorPrefix, ctx)
	if extraContext == nil {
		response.ErrorSystem(ctx, "json 化失败", "")
		return
	} else {
		(&admin.Topic{}).Create(extraContext)
	}
}

type AdminTopicList struct {
	common_data_type.Page
}

func (l AdminTopicList) CheckParams(ctx *gin.Context) {
	l.Page.SetDefault()
	l.Page.Limit = 100
	if err := ctx.ShouldBind(&l); err != nil {
		response.ValidatorError(ctx, err)
		return
	}

	extraContext := data_transfer.DataAddContext(l, consts.ValidatorPrefix, ctx)
	if extraContext == nil {
		response.ErrorSystem(ctx, "json 化失败", "")
		return
	} else {
		(&api.Topic{}).GetListByCategoryAndShowType(extraContext)
	}
}

type AdminTopicUpdate struct {
	ID    int64  `form:"id" json:"id" binding:"required,min=1"`
	Title string `form:"title" json:"title" binding:"required,min=1"`
	Body  string `form:"body" json:"body" binding:"required,min=1"`
}

func (l AdminTopicUpdate) CheckParams(ctx *gin.Context) {
	if err := ctx.ShouldBind(&l); err != nil {
		response.ValidatorError(ctx, err)
		return
	}

	extraContext := data_transfer.DataAddContext(l, consts.ValidatorPrefix, ctx)
	if extraContext == nil {
		response.ErrorSystem(ctx, "json 化失败", "")
		return
	} else {
		(&admin.Topic{}).Update(extraContext)
	}
}

type AdminTopicDelete struct {
	ID int64 `form:"id" json:"id" binding:"required,min=1"`
}

func (l AdminTopicDelete) CheckParams(ctx *gin.Context) {
	if err := ctx.ShouldBind(&l); err != nil {
		response.ValidatorError(ctx, err)
		return
	}

	extraContext := data_transfer.DataAddContext(l, consts.ValidatorPrefix, ctx)
	if extraContext == nil {
		response.ErrorSystem(ctx, "json 化失败", "")
		return
	} else {
		(&admin.Topic{}).Delete(extraContext)
	}
}

type AdminCustomerList struct {
	common_data_type.Page
}

func (l AdminCustomerList) CheckParams(ctx *gin.Context) {
	l.Page.SetDefault()
	if err := ctx.ShouldBind(&l); err != nil {
		response.ValidatorError(ctx, err)
		return
	}

	extraContext := data_transfer.DataAddContext(l, consts.ValidatorPrefix, ctx)
	if extraContext == nil {
		response.ErrorSystem(ctx, "json 化失败", "")
		return
	} else {
		(&admin.Customer{}).GetList(extraContext)
	}
}
