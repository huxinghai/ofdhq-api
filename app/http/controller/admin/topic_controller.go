package admin

import (
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/service/topic"
	"ofdhq-api/app/utils/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Topic struct {
}

func (t *Topic) GetListByCategoryAndShowType(ctx *gin.Context) {
	page := int(ctx.GetFloat64(consts.ValidatorPrefix + "page"))
	pageSize := int(ctx.GetFloat64(consts.ValidatorPrefix + "limit"))

	fact := topic.CreateTopicFactory()

	list, err := fact.GetAllList(page, pageSize)
	if err != nil {
		variable.ZapLog.Error("查询数据失败", zap.Error(err))
		response.Fail(ctx, consts.TopicGetListErrorCode, consts.TopicGetListErrorMsg, "")
		return
	}
	count, err := fact.GetCount()
	if err != nil {
		variable.ZapLog.Error("查询数据失败", zap.Error(err))
		response.Fail(ctx, consts.TopicGetListErrorCode, consts.TopicGetListErrorMsg, "")
		return
	}

	pageInfo := response.GenPageInfo(page, pageSize, int(count))
	response.Success(ctx, consts.CurdStatusOkMsg, gin.H{"page_info": pageInfo, "counts": count, "list": list})
}

func (t *Topic) Create(ctx *gin.Context) {
	title := ctx.GetString(consts.ValidatorPrefix + "title")
	body := ctx.GetString(consts.ValidatorPrefix + "body")
	adminUser, err := (&AdminUsers{}).buildCurrentUser(ctx)
	if err != nil {
		response.Fail(ctx, consts.TopicCreatedErrorCode, consts.TopicCreatedErrorMsg, "")
		return
	}
	id, err := topic.CreateTopicFactory().Create(title, body, adminUser.Id)
	if err != nil {
		response.Fail(ctx, consts.TopicCreatedErrorCode, consts.TopicCreatedErrorMsg, "")
		return
	}
	response.Success(ctx, consts.CurdStatusOkMsg, gin.H{"id": id})
}

func (t *Topic) Update(ctx *gin.Context) {
	id := int64(ctx.GetFloat64(consts.ValidatorPrefix + "id"))
	title := ctx.GetString(consts.ValidatorPrefix + "title")
	body := ctx.GetString(consts.ValidatorPrefix + "body")

	_, err := topic.CreateTopicFactory().Update(id, title, body)
	if err != nil {
		response.Fail(ctx, consts.TopicCreatedErrorCode, consts.TopicCreatedErrorMsg, "")
		return
	}
	response.Success(ctx, consts.CurdStatusOkMsg, gin.H{"id": id})
}

func (t *Topic) Delete(ctx *gin.Context) {
	id := int64(ctx.GetFloat64(consts.ValidatorPrefix + "id"))
	err := topic.CreateTopicFactory().DeleteTopicByID(id)
	if err != nil {
		response.Fail(ctx, consts.TopicDeleteErrorCode, err.Error(), "")
		return
	}
	response.Success(ctx, consts.CurdStatusOkMsg, gin.H{"id": id})
}
