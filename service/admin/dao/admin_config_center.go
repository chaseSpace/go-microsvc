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
	"time"
)

type adminConfigCenterDao struct {
}

var AdminConfigCenterDao adminConfigCenterDao

func (adminConfigCenterDao) Add(ctx context.Context, item *commonpb.ConfigItem) error {
	core := item.Core
	return xerr.WrapMySQL(admin.QAdmin.WithContext(ctx).Create(&admin.AdminConfigCenter{
		Key:            core.Key,
		Name:           core.Name,
		Value:          core.Value,
		IsLock:         core.IsLock,
		FieldCreatedBy: model.FieldCreatedBy{CreatedBy: item.CreatedBy},
		FieldUpdatedBy: model.FieldUpdatedBy{UpdatedBy: item.UpdatedBy},
	}).Error)
}

func (a adminConfigCenterDao) Delete(ctx context.Context, key string, caller *auth.AdminCaller) error {
	list, err := a.GetConfigByKeys(ctx, key)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return xerr.ErrDataNotExist.New("配置项不存在")
	}
	if list[0].IsLock && !caller.IsSuper && caller.Uid != list[0].CreatedBy {
		return xerr.ErrForbidden.New("配置项锁定，仅创建人或管理员可删除")
	}
	return xerr.WrapMySQL(admin.QAdmin.WithContext(ctx).Model(admin.AdminConfigCenter{}).
		Where("`key` = ?", key).
		Update("deleted_ts", time.Now().Unix()).Error)
}

func (a adminConfigCenterDao) UpdateConfigValue(ctx context.Context, item *commonpb.ConfigItemCore, byProgram bool, caller *auth.AdminCaller) error {
	list, err := a.GetConfigByKeys(ctx, item.Key)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return xerr.ErrDataNotExist.New("配置项不存在")
	}
	if list[0].IsLock && !caller.IsSuper && caller.Uid != list[0].CreatedBy {
		return xerr.ErrForbidden.New("配置项锁定，仅创建人或管理员可修改")
	}
	if item.IsLock && !list[0].IsLock && !caller.IsSuper && caller.Uid != list[0].CreatedBy {
		return xerr.ErrForbidden.New("仅创建人或管理员可锁定配置项")
	}
	if byProgram && !list[0].AllowProgramUpdate {
		return xerr.ErrParams.New("配置项不允许程序修改")
	}
	if byProgram && list[0].AllowProgramUpdate != item.AllowProgramUpdate {
		return xerr.ErrParams.New("属性{AllowProgramUpdate}不允许程序修改")
	}
	if byProgram && list[0].IsLock != item.IsLock {
		return xerr.ErrParams.New("属性{IsLock}不允许程序修改")
	}
	upMap := map[string]interface{}{
		"value":                item.Value,
		"name":                 item.Name,
		"updated_by":           caller.Uid,
		"is_lock":              item.IsLock,
		"allow_program_update": item.AllowProgramUpdate,
	}

	err = admin.QAdmin.WithContext(ctx).Model(&admin.AdminConfigCenter{}).Where("`key` = ?", item.Key).Updates(upMap).Error
	return xerr.WrapMySQL(err)
}

func (adminConfigCenterDao) GetConfigByKeys(ctx context.Context, key ...string) (list []*admin.AdminConfigCenter, err error) {
	err = admin.QAdmin.WithContext(ctx).Where("deleted_ts = 0").Find(&list, "`key` in (?)", key).Error
	err = xerr.WrapMySQL(err)
	return
}

func (adminConfigCenterDao) List(ctx context.Context, req *adminpb.ConfigCenterListReq) (list []*admin.AdminConfigCenter, total int64, err error) {
	q := admin.QAdmin.WithContext(ctx).Model(&admin.AdminConfigCenter{})
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
