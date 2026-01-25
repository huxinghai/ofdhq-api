package api

import (
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/http/controller/api/models"
	"ofdhq-api/app/service/topic"
	"ofdhq-api/app/utils/response"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/russross/blackfriday/v2"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

type Topic struct {
}

func (t *Topic) GetListByCategoryAndShowType(ctx *gin.Context) {
	lang := ctx.GetString(consts.ValidatorPrefix + "lang")
	page := int(ctx.GetFloat64(consts.ValidatorPrefix + "page"))
	pageSize := int(ctx.GetFloat64(consts.ValidatorPrefix + "limit"))

	list, err := topic.CreateTopicFactory().GetAllListByLang(lang, page, pageSize)
	if err != nil {
		variable.ZapLog.Error("查询数据失败", zap.Error(err))
		response.Fail(ctx, consts.TopicGetListErrorCode, consts.TopicGetListErrorMsg, "")
		return
	}

	totalCount, err := topic.CreateTopicFactory().GetCountByLang(lang)
	if err != nil {
		variable.ZapLog.Error("查询数据失败", zap.Error(err))
		response.Fail(ctx, consts.TopicGetListErrorCode, consts.TopicGetListErrorMsg, "")
		return
	}

	result := make([]*models.Topic, 0, len(list))
	for _, l := range list {
		html := string(blackfriday.Run([]byte(l.Body)))

		description := extractText(html)

		result = append(result, &models.Topic{
			Topic: &models.TopicBasic{
				ID:          l.Id,
				Lang:        l.Lang,
				Title:       l.Title,
				Body:        l.Body,
				BodyHtml:    html,
				Description: description,
				ImgUrl:      l.ImgUrl,
				Flag:        l.Flag,
				CreatedAt:   l.CreatedAt,
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
		html := string(blackfriday.Run([]byte(topic.Body)))
		result := &models.Topic{
			Topic: &models.TopicBasic{
				ID:          topic.Id,
				Title:       topic.Title,
				Body:        topic.Body,
				BodyHtml:    html,
				ImgUrl:      topic.ImgUrl,
				Description: "",
				Flag:        topic.Flag,
				CreatedAt:   topic.CreatedAt,
			},
			User: &models.User{
				ID:        topic.AdminUserID,
				Name:      "Haq",
				AvatarUrl: variable.DefaultAvatar,
			},
		}
		result.Topic.Description = extractText(result.Topic.BodyHtml)
		response.Success(ctx, consts.CurdStatusOkMsg, result)
	}
}

func extractText(htmlStr string) string {
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return ""
	}

	var buf strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return strings.Join(strings.Fields(buf.String()), " ")
}
