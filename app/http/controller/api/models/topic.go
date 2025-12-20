package models

import "ofdhq-api/app/model"

type Topic struct {
	Topic *model.TopicModel `json:"topic"`
	User  *User             `json:"user"`
}

type User struct {
	Name      string `json:"name"`
	ID        int64  `json:"id"`
	AvatarUrl string `json:"avatar_url"`
}
