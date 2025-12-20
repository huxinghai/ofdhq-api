package topic

import (
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/http/controller/api"
	common_data_type "ofdhq-api/app/http/validator/common/data_type"
	"ofdhq-api/app/http/validator/core/data_transfer"
	"ofdhq-api/app/utils/response"

	"github.com/gin-gonic/gin"
)

type List struct {
	common_data_type.Page
}

func (l List) CheckParams(ctx *gin.Context) {
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
		(&api.Topic{}).GetListByCategoryAndShowType(extraContext)
	}
}

type Detail struct {
	ID int64 `form:"id" json:"id" binding:"required,min=1"`
}

func (l Detail) CheckParams(ctx *gin.Context) {
	if err := ctx.ShouldBind(&l); err != nil {
		response.ValidatorError(ctx, err)
		return
	}

	extraContext := data_transfer.DataAddContext(l, consts.ValidatorPrefix, ctx)
	if extraContext == nil {
		response.ErrorSystem(ctx, "json 化失败", "")
		return
	} else {
		(&api.Topic{}).Detail(extraContext)
	}
}
