package dao

import (
	"context"
	"microsvc/enums"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"time"

	"gorm.io/gorm"
)

func CreateUserWeixin(tx *gorm.DB, ent *user.UserRegisterWeixin) error {
	ent.CreatedAt = time.Now()
	ent.UpdatedAt = ent.CreatedAt
	return tx.Create(ent).Error
}

func GetUserWeixin(ctx context.Context, openid string, typ enums.UserWxType) (t user.UserRegisterWeixin, err error) {
	err = user.Q.WithContext(ctx).Model(t).Where("account = ? and type=?", openid, typ).First(&t).Error
	return t, xerr.WrapMySQL(err)
}
