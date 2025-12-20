package model

import (
	"errors"
	"fmt"
	"ofdhq-api/app/global/variable"
	"ofdhq-api/app/utils/md5_encrypt"
	"time"

	"go.uber.org/zap"
)

func CreateAdminUserFactory() *AdminUsers {
	return &AdminUsers{BaseModel: BaseModel{DB: UseDbConn("")}}
}

func NewAdminUserFactory(userModel *AdminUsers) *AdminUsers {
	if userModel == nil {
		return CreateAdminUserFactory()
	} else {
		userModel.BaseModel.DB = UseDbConn("")
		return userModel
	}
}

type AdminUsers struct {
	BaseModel
	Pass          string `gorm:"column:pass" json:"-"`
	RealName      string `gorm:"column:real_name" json:"real_name"`
	Email         string `json:"email"`
	Status        int64  `json:"status"`
	RoleType      int64  `json:"role_type"` // 1-管理员，2-操作员
	LastLoginTime string `json:"last_login_time"`
	LastLoginIp   string `gorm:"column:last_login_ip" json:"last_login_ip"`
}

// 表名
func (u *AdminUsers) TableName() string {
	return "admin_users"
}

// 用户注册（写一个最简单的使用账号、密码注册即可）
func (u *AdminUsers) Register(realName, email, pass, userIp string, role int64) bool {
	sql := "INSERT INTO admin_users(real_name,email,pass,last_login_ip,role_type) VALUES(?,?,?,?,?)"
	result := u.Exec(sql, realName, email, pass, userIp, role)
	if result.RowsAffected > 0 {
		return true
	} else {
		return false
	}
}

func (u *AdminUsers) Update() error {
	if u.Email == "" || u.Id <= 0 {
		return fmt.Errorf("参数不能为空！")
	}
	sqlstr := `UPDATE admin_users SET real_name=?,email=?,role_type=? where id=?`
	tt := u.Exec(sqlstr, u.RealName, u.Email, u.RoleType, u.Id)
	if tt.Error != nil {
		return errors.Join(tt.Error, fmt.Errorf("更新失败！%+v", u))
	}
	return nil
}

func (u *AdminUsers) UpdatePassword() error {
	if u.Pass == "" || u.Id <= 0 {
		return fmt.Errorf("参数不能为空！")
	}
	sqlstr := `UPDATE admin_users SET pass=? where id=?`
	tt := u.Exec(sqlstr, u.Pass, u.Id)
	if tt.Error != nil {
		return errors.Join(tt.Error, fmt.Errorf("更新失败！%+v", u))
	}
	return nil
}

var adminUserColumnTemplate = "`id`,`pass`,`real_name`, `status`,`email`,`role_type`, last_login_ip, last_login_time,`created_at`"

// 用户登录,
func (u *AdminUsers) Login(email string, pass string) *AdminUsers {
	sqlStr := "select " + adminUserColumnTemplate + " from admin_users where email=? limit 1;"
	result := u.Raw(sqlStr, email).First(u)
	crpyt := md5_encrypt.Base64Md5(pass)
	fmt.Println(crpyt)
	if result.Error == nil {
		// 账号密码验证成功
		if len(u.Pass) > 0 && (u.Pass == md5_encrypt.Base64Md5(pass)) {
			return u
		}
	} else {
		variable.ZapLog.Error("根据Admin账号查询单条记录出错:", zap.Error(result.Error))
	}
	return nil
}

func (u *AdminUsers) List(page, limit int64) ([]*AdminUsers, error) {
	limitStart := (page - 1) * limit
	sql := "SELECT  " + adminUserColumnTemplate + " FROM  `admin_users` LIMIT ?,?"

	tmp := make([]*AdminUsers, 0)
	result := u.Raw(sql, limitStart, limit).Find(&tmp)
	if result.Error == nil {
		return tmp, nil
	} else {
		return nil, errors.Join(result.Error, fmt.Errorf("List 查询失败"))
	}
}

func (u *AdminUsers) GetByEmail(email string) (*AdminUsers, error) {
	sql := "SELECT  " + adminUserColumnTemplate + " FROM  `admin_users` WHERE email=? LIMIT 1"
	result := u.Raw(sql, email).First(u)
	if result.Error == nil {
		if result.RowsAffected <= 0 {
			return nil, nil
		}
		return u, nil
	} else {
		return nil, errors.Join(result.Error, fmt.Errorf("GetByEmail 查询失败"))
	}
}

// 根据用户ID查询一条信息
func (u *AdminUsers) GetByAdminUserID(adminUserID int64) (*AdminUsers, error) {
	sql := "SELECT  " + adminUserColumnTemplate + " FROM  `admin_users` WHERE id=? LIMIT 1"
	result := u.Raw(sql, adminUserID).First(u)
	if result.Error == nil {
		if result.RowsAffected <= 0 {
			return nil, nil
		}
		return u, nil
	} else {
		return nil, errors.Join(result.Error, fmt.Errorf("GetByAdminUserID 查询失败"))
	}
}

// 记录用户登陆（login）生成的token，每次登陆记录一次token
func (u *AdminUsers) OauthLoginToken(adminUserId int64, token string, expiresAt int64, clientIp string) bool {
	sql := `
		INSERT INTO admin_oauth_access_tokens(admin_user_id,action_name,token,expires_at,client_ip)
		SELECT  ?,'login',? ,?,? FROM DUAL WHERE NOT EXISTS(SELECT 1 FROM  admin_oauth_access_tokens a WHERE  a.admin_user_id=?  AND a.action_name='login' AND a.token=?)
	`
	//注意：token的精确度为秒，如果在一秒之内，一个账号多次调用接口生成的token其实是相同的，这样写入数据库，第二次的影响行数为0，知己实际上操作仍然是有效的。
	//所以这里只判断无错误即可，判断影响行数的话，>=0 都是ok的
	if u.Exec(sql, adminUserId, token, time.Unix(expiresAt, 0).Format(variable.DateFormat), clientIp, adminUserId, token).Error == nil {
		return true
	}
	return false
}

// 用户刷新token,条件检查: 相关token在过期的时间之内，就符合刷新条件
func (u *AdminUsers) OauthRefreshConditionCheck(adminUserId int64, oldToken string) bool {
	// 首先判断旧token在本系统自带的数据库已经存在，才允许继续执行刷新逻辑
	var oldTokenIsExists int
	sql := "SELECT count(*)  as counts FROM admin_oauth_access_tokens  WHERE admin_user_id =? and token=? and NOW()<DATE_ADD(expires_at,INTERVAL ? SECOND)"
	if u.Raw(sql, adminUserId, oldToken, variable.ConfigYml.GetInt64("Token.JwtTokenRefreshAllowSec")).First(&oldTokenIsExists).Error == nil && oldTokenIsExists == 1 {
		return true
	}
	return false
}

// 更新用户登陆次数、最近一次登录ip、最近一次登录时间
func (u *AdminUsers) UpdateUserloginInfo(last_login_ip string, adminUserId int64) {
	sql := "UPDATE admin_users SET login_times=IFNULL(login_times,0)+1,last_login_ip=?,last_login_time=?  WHERE  id=?  "
	_ = u.Exec(sql, last_login_ip, variable.NowTimeSH().Format(variable.DateFormat), adminUserId)
}

// 当用户更改密码后，所有的token都失效，必须重新登录
func (u *AdminUsers) OauthResetToken(adminUserId int64, newPass, clientIp string) bool {
	//如果用户新旧密码一致，直接返回true，不需要处理
	userItem, err := u.GetByAdminUserID(adminUserId)
	if userItem != nil && err == nil && userItem.Pass == newPass {
		return true
	} else if userItem != nil {
		sql := "UPDATE admin_oauth_access_tokens  SET  revoked=1,updated_at=NOW(),action_name='ResetPass',client_ip=?  WHERE  admin_user_id=?  "
		if u.Exec(sql, clientIp, adminUserId).Error == nil {
			return true
		}
	}
	return false
}

// 用户刷新token
func (u *AdminUsers) OauthRefreshToken(adminUserId, expiresAt int64, oldToken, newToken, clientIp string) bool {
	sql := "UPDATE  admin_oauth_access_tokens SET  token=? ,expires_at=?,client_ip=?,updated_at=NOW(),action_name='refresh'  WHERE admin_user_id=? AND token=?"
	if u.Exec(sql, newToken, time.Unix(expiresAt, 0).Format(variable.DateFormat), clientIp, adminUserId, oldToken).Error == nil {
		return true
	}
	return false
}

// 判断用户token是否在数据库存在+状态OK
func (u *AdminUsers) OauthCheckTokenIsOk(adminUserId int64, token string) bool {
	sql := "SELECT  token  FROM  `admin_oauth_access_tokens`  WHERE   admin_user_id=?  AND  revoked=0  AND  expires_at>NOW() ORDER  BY  expires_at  DESC , updated_at  DESC  LIMIT ?"
	maxOnlineUsers := variable.ConfigYml.GetInt("Token.JwtTokenOnlineUsers")
	rows, err := u.Raw(sql, adminUserId, maxOnlineUsers).Rows()
	defer func() {
		//  凡是查询类记得释放记录集
		_ = rows.Close()
	}()
	if err == nil && rows != nil {
		for rows.Next() {
			var tempToken string
			err := rows.Scan(&tempToken)
			if err == nil {
				if tempToken == token {
					return true
				}
			}
		}
	}
	return false
}
