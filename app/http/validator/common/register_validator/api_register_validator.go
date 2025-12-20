package register_validator

import (
	"ofdhq-api/app/core/container"
	"ofdhq-api/app/global/consts"
	"ofdhq-api/app/http/validator/admins"
	"ofdhq-api/app/http/validator/api/topic"
)

// 各个业务模块验证器必须进行注册（初始化），程序启动时会自动加载到容器
func ApiRegisterValidator() {
	//创建容器
	containers := container.CreateContainersFactory()

	//  key 按照前缀+模块+验证动作 格式，将各个模块验证注册在容器
	var key string

	// topic
	key = consts.ValidatorPrefix + "TopicList"
	containers.Set(key, topic.List{})
	key = consts.ValidatorPrefix + "TopicDetail"
	containers.Set(key, topic.Detail{})

	//==========================================================  管理后台
	key = consts.ValidatorPrefix + "AdminUserLogin"
	containers.Set(key, admins.UserLogin{})
	key = consts.ValidatorPrefix + "AdminUserRegister"
	containers.Set(key, admins.UserRegister{})
	key = consts.ValidatorPrefix + "AdminAdminUserUpdate"
	containers.Set(key, admins.UserUpdate{})
	key = consts.ValidatorPrefix + "AdminUserUpdatePassword"
	containers.Set(key, admins.UserUpdatePassword{})
	key = consts.ValidatorPrefix + "AdminAdminUserList"
	containers.Set(key, admins.AdminUserList{})

	key = consts.ValidatorPrefix + "AdminTopicCreate"
	containers.Set(key, admins.AdminTopicCreate{})
	key = consts.ValidatorPrefix + "AdminTopicDelete"
	containers.Set(key, admins.AdminTopicDelete{})
	key = consts.ValidatorPrefix + "AdminTopicUpdate"
	containers.Set(key, admins.AdminTopicUpdate{})
	key = consts.ValidatorPrefix + "AdminTopicList"
	containers.Set(key, admins.AdminTopicList{})
}
