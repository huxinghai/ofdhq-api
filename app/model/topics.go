package model

import (
	"errors"
	"fmt"
	"ofdhq-api/app/global/consts"
)

func CreateTopicFactory() *TopicModel {
	return &TopicModel{
		BaseModel: BaseModel{DB: UseDbConn("")},
	}
}

func NewTopicFactory(sc *TopicModel) *TopicModel {
	if sc == nil {
		return CreateTopicFactory()
	} else {
		sc.BaseModel.DB = UseDbConn("")
		return sc
	}
}

type TopicModel struct {
	BaseModel
	AdminUserID int64  `json:"admin_user_id"` // 作者 ID
	Title       string `json:"title"`         // 标题
	Body        string `json:"body"`          // 内容
	ImgUrl      string `json:"img_url"`
	Flag        int    `json:"flag"` // 状态：0-无效，1-有效
}

func (t *TopicModel) TableName() string {
	return "topics"
}

func (t *TopicModel) Insert() error {
	if t.Title == "" {
		return errors.New("TopicModel.Create 没有Title参数!")
	}
	if t.Body == "" {
		return errors.New("TopicModel.Create 没有Body参数!")
	}
	tt := t.DB.Create(t)
	if tt.Error != nil {
		return errors.Join(tt.Error, fmt.Errorf("TopicModel.Create 失败"))
	}
	return nil
}

func (t *TopicModel) Update() error {
	if t.Id <= 0 {
		return errors.New("ID 为空!")
	}
	if t.Title == "" {
		return errors.New("TopicModel.Create 没有Title参数!")
	}
	if t.Body == "" {
		return errors.New("TopicModel.Create 没有Body参数!")
	}
	tt := t.DB.Save(t)
	if tt.Error != nil {
		return errors.Join(tt.Error, fmt.Errorf("TopicModel.Create 失败"))
	}
	return nil
}

func (t *TopicModel) DeleteById(id int64) error {
	sqlstr := "UPDATE `topics` SET flag = ? WHERE `id`= ?"
	result := t.Exec(sqlstr, consts.DBFlagFieldInvalid, id)
	if result.Error == nil {
		return nil
	} else {
		return errors.Join(result.Error, fmt.Errorf("TopicModel.DeleteById 删除失败"))
	}
}

func (t *TopicModel) UpdateById(id int64, title, body string) error {
	sqlstr := "UPDATE `topics` SET `title`= ?, `body`= ? WHERE `id`= ?"
	result := t.Exec(sqlstr, title, body, id)
	if result.Error == nil {
		return nil
	} else {
		return errors.Join(result.Error, fmt.Errorf("TopicModel.UpdateById 更新失败"))
	}
}

var topicColumnTemplate = "`id`, `admin_user_id`, `title`, `body`, `img_url`,`created_at`, `updated_at`"

func (t *TopicModel) GetAll(page, limit int) ([]*TopicModel, error) {
	limitStart := (page - 1) * limit
	sqlstr := "SELECT " + topicColumnTemplate + " FROM `topics` WHERE flag=1 order by created_at desc LIMIT ?,?"
	tmp := make([]*TopicModel, 0)
	result := t.Raw(sqlstr, limitStart, limit).Find(&tmp)
	if result.Error == nil {
		return tmp, nil
	} else {
		return nil, errors.Join(result.Error, fmt.Errorf("TopicModel.GetAll 查询失败"))
	}
}

func (t *TopicModel) GetCount() (int64, error) {
	sqlstr := "SELECT count(0) FROM `topics` WHERE flag=1"
	var count int64
	result := t.Raw(sqlstr).First(&count)
	if result.Error == nil {
		return count, nil
	} else {
		return 0, errors.Join(result.Error, fmt.Errorf("TopicModel.GetCount 查询失败"))
	}
}

func (t *TopicModel) GetById(id int64) (*TopicModel, error) {
	sqlstr := "SELECT " + topicColumnTemplate + " FROM `topics` WHERE flag=1 and id=?"
	tmp := &TopicModel{}
	result := t.Raw(sqlstr, id).Find(tmp)
	if result.Error == nil {
		if result.RowsAffected <= 0 {
			return nil, nil
		}
		return tmp, nil
	} else {
		return nil, errors.Join(result.Error, fmt.Errorf("TopicModel.GetById 查询失败"))
	}
}

func (t *TopicModel) GetListByUserId(userId int64, page, limit int64) ([]*TopicModel, error) {
	limitStart := (page - 1) * limit
	sqlstr := "SELECT " + topicColumnTemplate + " FROM `topics` WHERE flag=1 and user_id=? ORDER BY created_at DESC LIMIT ?,?"
	tmp := make([]*TopicModel, 0)
	result := t.Raw(sqlstr, userId, limitStart, limit).Find(&tmp)
	if result.Error == nil {
		return tmp, nil
	} else {
		return nil, errors.Join(result.Error, fmt.Errorf("TopicModel.GetListByUserId 查询失败"))
	}
}

func (t *TopicModel) GetListCountByUserId(userId int64) (int64, error) {
	sqlstr := "SELECT count(0) FROM `topics` WHERE flag=1 and user_id=?"
	var count int64
	result := t.Raw(sqlstr, userId).First(&count)
	if result.Error == nil {
		return count, nil
	} else {
		return 0, errors.Join(result.Error, fmt.Errorf("TopicModel.GetListByUserId 查询失败"))
	}
}
