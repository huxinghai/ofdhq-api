package api

import (
	"bytes"
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/http/controller/api/models"
	"ofdhq-api/app/service/topic"
	"ofdhq-api/app/utils/response"

	"github.com/gin-gonic/gin"
	"github.com/yuin/goldmark"
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
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(l.Body), &buf); err != nil {
			response.Fail(ctx, consts.TopicGetListErrorCode, consts.TopicGetListErrorMsg, "渲染markdown失败")
			return
		}

		result = append(result, &models.Topic{
			Topic: &models.TopicBasic{
				ID:        l.Id,
				Title:     l.Title,
				Body:      l.Body,
				BodyHtml:  buf.String(),
				ImgUrl:    l.ImgUrl,
				Flag:      l.Flag,
				CreatedAt: l.CreatedAt,
			},
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
		if topic == nil {
			response.Fail(ctx, consts.TopicGetDetailErrorCode, consts.TopicGetDetailErrorMsg, "")
			return
		}
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(topic.Body), &buf); err != nil {
			response.Fail(ctx, consts.TopicGetListErrorCode, consts.TopicGetListErrorMsg, "渲染markdown失败")
			return
		}
		result := &models.Topic{
			Topic: &models.TopicBasic{
				ID:        topic.Id,
				Title:     topic.Title,
				Body:      topic.Body,
				BodyHtml:  buf.String(),
				ImgUrl:    topic.ImgUrl,
				Flag:      topic.Flag,
				CreatedAt: topic.CreatedAt,
			},
			User: &models.User{
				ID:        topic.AdminUserID,
				Name:      "Haq",
				AvatarUrl: variable.DefaultAvatar,
			},
		}
		response.Success(ctx, consts.CurdStatusOkMsg, result)
	}
}
