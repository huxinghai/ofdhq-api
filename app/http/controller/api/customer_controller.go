package api

import (
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/model"
	"ofdhq-api/app/utils/response"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Customers struct {
}

func (t *Customers) Create(ctx *gin.Context) {
	firstName := ctx.GetString(consts.ValidatorPrefix + "first_name")
	lastName := ctx.GetString(consts.ValidatorPrefix + "last_name")
	email := ctx.GetString(consts.ValidatorPrefix + "email")
	subject := ctx.GetString(consts.ValidatorPrefix + "subject")
	messages := ctx.GetString(consts.ValidatorPrefix + "messages")

	customer := &model.CustomerModel{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Subject:   subject,
		Messages:  messages,
	}
	createTime := time.Now().Format("2006-01-02 15:04:05")
	customer.CreatedAt = createTime
	customer.UpdatedAt = createTime
	tx := model.CreateCustomerFactory().Create(customer)
	if tx.Error != nil {
		variable.ZapLog.Error("创建客户联系方式失败", zap.Error(tx.Error))
		response.Fail(ctx, consts.TopicGetDetailErrorCode, consts.TopicGetDetailErrorMsg, "")
		return
	}

	response.Success(ctx, consts.CurdStatusOkMsg, gin.H{"id": customer.Id})
}
