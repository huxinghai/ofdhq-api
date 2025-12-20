package adminusers

import (
	"ofdhq-api/app/model"
	"ofdhq-api/app/utils/md5_encrypt"
)

func CreateAdminUserFactory() *AdminUsers {
	return &AdminUsers{model.CreateAdminUserFactory()}
}

func NewAdminUserFactory(adminUser *model.AdminUsers) *AdminUsers {
	return &AdminUsers{model.NewAdminUserFactory(adminUser)}
}

type AdminUsers struct {
	adminUsers *model.AdminUsers
}

func (u *AdminUsers) Register(realName, email, pass, userIp string, role int64) bool {
	pass = md5_encrypt.Base64Md5(pass) // 预先处理密码加密，然后存储在数据库
	return u.adminUsers.Register(realName, email, pass, userIp, role)
}
