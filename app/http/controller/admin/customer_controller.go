package admin

import (
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/model"
	"ofdhq-api/app/utils/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Customer struct {
}

func (t *Customer) GetList(ctx *gin.Context) {
	page := int(ctx.GetFloat64(consts.ValidatorPrefix + "page"))
	pageSize := int(ctx.GetFloat64(consts.ValidatorPrefix + "limit"))

	customerList, err := model.CreateCustomerFactory().GetAll(page, pageSize)
	if err != nil {
		variable.ZapLog.Error("CreateCustomerFactory().GetAll失败！", zap.Error(err))
		response.Fail(ctx, consts.ServerOccurredErrorCode, consts.ServerOccurredErrorMsg, "")
		return
	}
	count, err := model.CreateCustomerFactory().GetCount()
	if err != nil {
		variable.ZapLog.Error("CreateCustomerFactory().GetCount失败！", zap.Error(err))
		response.Fail(ctx, consts.ServerOccurredErrorCode, consts.ServerOccurredErrorMsg, "")
		return
	}
	pageInfo := response.GenPageInfo(page, pageSize, int(count))
	response.Success(ctx, consts.CurdStatusOkMsg, gin.H{"page_info": pageInfo, "counts": count, "list": customerList})
}
