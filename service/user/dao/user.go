package dao

import (
	"context"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"time"

	"gorm.io/gorm"
)

func CreateUser(tx *gorm.DB, ent *user.User) error {
	if err := ent.Check(); err != nil {
		return xerr.ErrInvalidRegisterInfo.AppendMsg(err.Error())
	}
	ent.CreatedAt = time.Now()
	ent.UpdatedAt = ent.CreatedAt
	//if tx == nil {
	//	return user.Q.Create(ent).Error
	//}
	return tx.Create(ent).Error
}

func CreateUserExt(tx *gorm.DB, ent *user.UserExt) error {
	//if tx == nil {
	//	return user.Q.Create(ent).Error
	//}
	return tx.Create(ent).Error
}

func GetMaxUid(ctx context.Context) (uint64, error) {
	row := new(user.User)
	err := user.Q.WithContext(ctx).Order("uid desc").Take(row).Error
	return uint64(row.Uid), xerr.WrapMySQL(err)
}

func GetUser(ctx context.Context, uid ...int64) (list []*user.User, row *user.User, err error) {
	err = user.Q.WithContext(ctx).Find(&list, "uid in (?) or nid in (?)", uid, uid).Error
	err = xerr.WrapMySQL(err)
	if err != nil {
		return
	}
	if len(uid) == 1 && len(list) == 1 {
		row = list[0]
	}
	return
}

func GetUserByPhone(ctx context.Context, in ...string) (list []*user.User, row user.User, err error) {
	err = user.Q.WithContext(ctx).Find(&list, "phone in (?)", in).Error
	if len(list) > 0 {
		row = *list[0]
	}
	err = xerr.WrapMySQL(err)
	return
}

func GetUserFromTh(ctx context.Context, type_ commonpb.SignInType, accounts ...string) (list []*user.UserRegisterTh, row user.UserRegisterTh, err error) {
	err = user.Q.WithContext(ctx).Find(&list, "account in (?) and th_type = ?",
		accounts, type_).Error
	if len(list) > 0 {
		row = *list[0]
	}
	err = xerr.WrapMySQL(err)
	return
}

func UpdateUserNid(ctx context.Context, uid int64, nid *int64) error {
	return user.Q.WithContext(ctx).Model(&user.User{}).Where("uid=?", uid).Update("nid", nid).Error
}
