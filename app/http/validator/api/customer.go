package api

import (
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/http/controller/api"
	"ofdhq-api/app/http/validator/core/data_transfer"
	"ofdhq-api/app/utils/response"

	"github.com/gin-gonic/gin"
)

type Customer struct {
	FirstName string `form:"first_name" json:"first_name" binding:"required,min=1"`
	LastName  string `form:"last_name" json:"last_name" binding:"required,min=1"`
	Email     string `form:"email" json:"email" binding:"required,min=1"`
	Subject   string `form:"subject" json:"subject" binding:"required,min=1"`
	Messages  string `form:"messages" json:"messages"`
}

func (l Customer) CheckParams(ctx *gin.Context) {
	if err := ctx.ShouldBind(&l); err != nil {
		response.ValidatorError(ctx, err)
		return
	}

	extraContext := data_transfer.DataAddContext(l, consts.ValidatorPrefix, ctx)
	if extraContext == nil {
		response.ErrorSystem(ctx, "json 化失败", "")
		return
	} else {
		(&api.Customers{}).Create(extraContext)
	}
}
