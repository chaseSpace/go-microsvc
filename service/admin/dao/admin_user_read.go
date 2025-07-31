package dao

import (
	"context"
	"microsvc/model/svc/admin"
	"microsvc/pkg/xerr"
)

func GetUser(ctx context.Context, uid ...int64) (list []*admin.AdminUser, row admin.AdminUser, err error) {
	if len(uid) == 1 {
		err = admin.QAdmin.WithContext(ctx).Take(&row, "uid = ?", uid[0]).Error
	} else {
		err = admin.QAdmin.WithContext(ctx).Find(&list, "uid in (?)", uid).Error
		if len(list) > 0 {
			row = *list[0]
		}
	}
	err = xerr.WrapMySQL(err)
	return
}

func GetUserByNickname(ctx context.Context, name ...string) (list []*admin.AdminUser, row admin.AdminUser, err error) {
	if len(name) == 1 {
		err = admin.QAdmin.WithContext(ctx).Take(&row, "nickname=?", name[0]).Error
	} else {
		err = admin.QAdmin.WithContext(ctx).Find(&list, "nickname in (?)", name).Error
	}
	err = xerr.WrapMySQL(err)
	return
}

func GetUserByPhone(ctx context.Context, phone ...string) (list []*admin.AdminUser, row admin.AdminUser, err error) {
	if len(phone) == 1 {
		err = admin.QAdmin.WithContext(ctx).Take(&row, "phone=?", phone[0]).Error
	} else {
		err = admin.QAdmin.WithContext(ctx).Find(&list, "phone in (?)", phone).Error
	}
	err = xerr.WrapMySQL(err)
	return
}
