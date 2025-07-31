package admin

import (
	"microsvc/enums"
	"microsvc/model"
	"time"
)

const (
	TableUKAdminUserPhone    = "'uk_phone'"
	TableUKAdminUserNickname = "'uk_nickname'"
)

/*
AdminUser
- 手机管理员账号采用 授权注册制
- 昵称账号由超管创建，并带有初始密码
*/
type AdminUser struct {
	model.FieldAt
	Uid         int64     `gorm:"column:uid" json:"uid"` // 内部id
	Icon        string    `gorm:"column:icon" json:"icon"`
	Nickname    string    `gorm:"column:nickname" json:"nickname"` // 唯一，可用于登录（忽略大小写）
	Description string    `gorm:"column:description" json:"description"`
	Birthday    time.Time `gorm:"column:birthday" json:"birthday"` // DB类型：date
	Sex         enums.Sex `gorm:"column:sex" json:"sex"`
	PasswdSalt  string    `gorm:"column:password_salt" json:"password_salt"`
	Password    string    `gorm:"column:password" json:"password"`
	Phone       *string   `gorm:"column:phone" json:"phone"`
}

func (u *AdminUser) TableName() string {
	return "admin_user"
}
