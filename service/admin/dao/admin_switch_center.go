package dao

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/model"
	"microsvc/model/svc/admin"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/commonpb"
	"microsvc/util/db"
	"microsvc/util/ujson"
	"time"
)

type adminSwitchCenterDao struct {
}

var AdminSwitchCenterDao adminSwitchCenterDao

func (adminSwitchCenterDao) Add(ctx context.Context, item *commonpb.SwitchItem) error {
	core := item.Core
	return xerr.WrapMySQL(admin.QAdmin.WithContext(ctx).Create(&admin.AdminSwitchCenter{
		Key:            core.Key,
		Name:           core.Name,
		Value:          core.Value,
		ValueExt:       core.ValueExt,
		IsLock:         core.IsLock,
		FieldCreatedBy: model.FieldCreatedBy{CreatedBy: item.CreatedBy},
		FieldUpdatedBy: model.FieldUpdatedBy{UpdatedBy: item.UpdatedBy},
	}).Error)
}

func (a adminSwitchCenterDao) Delete(ctx context.Context, key string, caller *auth.AdminCaller) error {
	list, err := a.GetSwitchByKeys(ctx, key)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return xerr.ErrDataNotExist
	}
	if list[0].IsLock && !caller.IsSuper && caller.Uid != list[0].CreatedBy {
		return xerr.ErrForbidden.New("开关项锁定，仅创建人或管理员可删除")
	}
	return xerr.WrapMySQL(admin.QAdmin.WithContext(ctx).Model(admin.AdminSwitchCenter{}).
		Where("`key` = ?", key).
		Update("deleted_ts", time.Now().Unix()).Error)
}

func (a adminSwitchCenterDao) UpdateSwitch(ctx context.Context, item *commonpb.SwitchItemCore, caller *auth.AdminCaller) error {
	list, err := a.GetSwitchByKeys(ctx, item.Key)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return xerr.ErrDataNotExist.New("开关项不存在")
	}
	if list[0].IsLock && !caller.IsSuper && caller.Uid != list[0].CreatedBy {
		return xerr.ErrForbidden.New("开关项锁定，仅创建人或管理员可修改")
	}
	if item.IsLock && !list[0].IsLock && !caller.IsSuper && caller.Uid != list[0].CreatedBy {
		return xerr.ErrForbidden.New("仅创建人或管理员可锁定开关")
	}
	if item.ValueExt == nil {
		item.ValueExt = make(admin.SwitchValExt)
	} else {
		err = admin.SwitchValExt(item.ValueExt).Check(item.Value)
		if err != nil {
			return err
		}
	}
	upMap := map[string]interface{}{
		"value":      item.Value,
		"name":       item.Name,
		"updated_by": caller.Uid,
		"is_lock":    item.IsLock,
		"value_ext":  ujson.MustMarshal2Str(item.ValueExt),
	}

	err = admin.QAdmin.WithContext(ctx).Model(&admin.AdminSwitchCenter{}).Where("`key` = ?", item.Key).Updates(upMap).Error
	return xerr.WrapMySQL(err)
}

func (adminSwitchCenterDao) GetSwitchByKeys(ctx context.Context, key ...string) (list []*admin.AdminSwitchCenter, err error) {
	err = admin.QAdmin.WithContext(ctx).Where("deleted_ts = 0").Find(&list, "`key` in (?)", key).Error
	err = xerr.WrapMySQL(err)
	return
}

func (adminSwitchCenterDao) List(ctx context.Context, req *adminpb.SwitchCenterListReq) (list []*admin.AdminSwitchCenter, total int64, err error) {
	q := admin.QAdmin.WithContext(ctx).Model(&admin.AdminSwitchCenter{})
	if req.Key != "" {
		err = q.Where("`key` = ?", req.Key).Find(&list).Error
		total = int64(len(list))
		return
	}

	if req.Name != "" {
		q = q.Where("name like ?", "%"+req.Name+"%")
	}
	q = q.Where("deleted_ts = 0")
	err = db.PageQuery(q, req.Page, "created_at desc, updated_at desc, id desc", &total, &list)
	err = xerr.WrapMySQL(err)
	return
}
