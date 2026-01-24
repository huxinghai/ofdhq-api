package topic

import (
	"errors"
	"fmt"
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/model"
	"regexp"
	"time"
)

type Service struct {
	TopicModel *model.TopicModel
}

func CreateTopicFactory() *Service {
	return &Service{TopicModel: model.CreateTopicFactory()}
}

func NewTopicFactory(sc *model.TopicModel) *Service {
	return &Service{TopicModel: model.NewTopicFactory(sc)}
}

// GetTopicByID
// 通过 ID 获取 topic 信息
func (t *Service) GetTopicByID(id int64) (*model.TopicModel, error) {
	return t.TopicModel.GetById(id)
}

// AddTopic
// 添加新的 tipic
func (t *Service) AddTopic(title, body string, userID, categoryID int64) error {
	return t.TopicModel.Insert()
}

// DeleteTopicByID
// 通过 ID 删除 topic
func (t *Service) DeleteTopicByID(id int64) error {
	return t.TopicModel.DeleteById(id)
}

func (t *Service) GetAllListByLang(lang string, page, limit int) ([]*model.TopicModel, error) {
	return t.TopicModel.GetAll(lang, page, limit)
}

func (t *Service) GetCountByLang(lang string) (int64, error) {
	return t.TopicModel.GetCount(lang)
}

func (t *Service) GetListByUser(userID int64, page, limit int64) (totalCount int64, list []*model.TopicModel, err error) {
	count, err := model.CreateTopicFactory().GetListCountByUserId(userID)
	if err != nil {
		return 0, nil, errors.Join(err, fmt.Errorf("GetListByUser 获取数据失败！userID:%d", userID))
	}
	list = make([]*model.TopicModel, 0)
	if count > 0 {
		list, err = model.CreateTopicFactory().GetListByUserId(userID, page, limit)
		if err != nil {
			return 0, nil, errors.Join(err, fmt.Errorf("GetListByUser 获取数据失败！userID:%d", userID))
		}
	}
	return count, list, nil
}

func (t *Service) Create(lang, title, body string, adminUserID int64) (int64, error) {
	topic := &model.TopicModel{
		Lang:        lang,
		Title:       title,
		Body:        body,
		AdminUserID: adminUserID,
		Flag:        consts.DBFlagFieldValid,
		ImgUrl:      extractFirstImageLink(body),
	}
	createTime := time.Now().Format("2006-01-02 15:04:05")
	topic.CreatedAt = createTime
	topic.UpdatedAt = createTime

	err := model.NewTopicFactory(topic).Insert()
	if err != nil {
		return 0, err
	}

	return topic.Id, nil
}

func (t *Service) Update(id int64, title, body string) (*model.TopicModel, error) {
	topic, err := model.CreateTopicFactory().GetById(id)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("查询数据失败！%d", id))
	}

	if topic == nil {
		return nil, errors.Join(err, fmt.Errorf("查询数据为空 %d", id))
	}

	topic.Title = title
	topic.Body = body
	topic.ImgUrl = extractFirstImageLink(body)
	topic.UpdatedAt = variable.NowTimeSH().Format(variable.DateFormat)
	topic.Flag = 1

	tt := model.NewTopicFactory(topic).Save(topic)
	if tt.Error != nil {
		return nil, errors.Join(tt.Error, fmt.Errorf("更新topic 失败！%+v", topic))
	}
	return topic, nil
}

func extractFirstImageLink(content string) string {
	firstImageLink := extractFirstImageLinkFromMarkdown(content)
	if firstImageLink != "" {
		return firstImageLink
	}
	return extractFirstImageLinkFromTxt(content)
}

// 从 markdown 中匹配第一张图片链接
func extractFirstImageLinkFromMarkdown(markdown string) string {
	// 正则表达式模式
	pattern := `!\[.*?\]\((.*?)\)`

	// 编译正则表达式
	regex := regexp.MustCompile(pattern)

	// 查找第一个匹配项
	match := regex.FindStringSubmatch(markdown)
	if match == nil {
		fmt.Println("No image link found")
		return ""
	}
	// 提取图片链接
	imageLink := match[1]
	return imageLink
}

// 从非markdown文本中匹配第一张图片链接
func extractFirstImageLinkFromTxt(text string) string {
	// 正则表达式模式
	pattern := `(?i)\bhttps?://\S+\.(?:png|jpe?g|gif)\b`

	// 编译正则表达式
	regex := regexp.MustCompile(pattern)

	// 查找第一个匹配项
	match := regex.FindString(text)
	if match == "" {
		return ""
	}

	return match
}
