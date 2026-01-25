package models

type Topic struct {
	Topic *TopicBasic `json:"topic"`
	User  *User       `json:"user"`
}

type User struct {
	Name      string `json:"name"`
	ID        int64  `json:"id"`
	AvatarUrl string `json:"avatar_url"`
}

type TopicBasic struct {
	ID          int64  `json:"id"`
	Lang        string `json:"lang"`
	Title       string `json:"title"` // 标题
	Body        string `json:"body"`  // 内容
	BodyHtml    string `json:"body_html"`
	Description string `json:"description"`
	ImgUrl      string `json:"img_url"`
	Flag        int    `json:"flag"` // 状态：0-无效，1-有效
	CreatedAt   string `json:"created_at"`
}
