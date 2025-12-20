package api

import (
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/http/controller/api/models"
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

	list, err := topic.CreateTopicFactory().GetAllList(page, pageSize)
	if err != nil {
		variable.ZapLog.Error("查询数据失败", zap.Error(err))
		response.Fail(ctx, consts.TopicGetListErrorCode, consts.TopicGetListErrorMsg, "")
		return
	}

	totalCount, err := topic.CreateTopicFactory().GetCount()
	if err != nil {
		variable.ZapLog.Error("查询数据失败", zap.Error(err))
		response.Fail(ctx, consts.TopicGetListErrorCode, consts.TopicGetListErrorMsg, "")
		return
	}

	result := make([]*models.Topic, 0, len(list))
	for _, l := range list {
		result = append(result, &models.Topic{
			Topic: l,
			User: &models.User{
				ID:        l.AdminUserID,
				Name:      "Haq",
				AvatarUrl: variable.DefaultAvatar,
			},
		})
	}

	pageInfo := response.GenPageInfo(page, pageSize, int(totalCount))
	response.Success(ctx, consts.CurdStatusOkMsg, gin.H{"page_info": pageInfo, "counts": totalCount, "list": result})
}

func (t *Topic) Detail(ctx *gin.Context) {
	id := int64(ctx.GetFloat64(consts.ValidatorPrefix + "id"))
	topic, err := topic.CreateTopicFactory().GetTopicByID(id)
	if err != nil {
		response.Fail(ctx, consts.TopicGetDetailErrorCode, consts.TopicGetDetailErrorMsg, "")
		return
	} else {
		result := &models.Topic{
			Topic: topic,
			User: &models.User{
				ID:        topic.AdminUserID,
				Name:      "Haq",
				AvatarUrl: variable.DefaultAvatar,
			},
		}
		response.Success(ctx, consts.CurdStatusOkMsg, result)
	}
}
