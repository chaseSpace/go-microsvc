package logic_misc

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/model/svc/admin"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/commonpb"
	"microsvc/service/admin/dao"
	"strings"

	"github.com/pkg/errors"

	"github.com/samber/lo"
)

type ctrl struct {
}

var Ext ctrl

func (ctrl) ConfigCenterAdd(ctx context.Context, req *adminpb.ConfigCenterAddReq) (*adminpb.ConfigCenterAddRes, error) {
	if req.Item == nil {
		return nil, xerr.ErrParams.New("请提供item")
	}
	caller := auth.ExtractAdminUser(ctx)
	core := req.Item
	core.Key = strings.ReplaceAll(core.Key, " ", "") // 过滤空格
	if core.Key == "" || core.Name == "" || core.Value == "" {
		return nil, xerr.ErrParams.New("请提供key/name/value")
	}
	item := &commonpb.ConfigItem{
		Core:      core,
		UpdatedBy: caller.Uid,
		CreatedBy: caller.Uid,
	}

	for i := 0; i < 2; i++ {
		err := dao.AdminConfigCenterDao.Add(ctx, item)
		if xerr.ErrDataDuplicate.Equal(err) {
			if req.IsOverride {
				err = dao.AdminConfigCenterDao.Delete(ctx, req.Item.Key, caller)
				if err != nil {
					return nil, err
				}
				continue
			}
			return nil, errors.New("配置Key已存在")
		}
		if err != nil {
			return nil, err
		}
		break
	}

	return &adminpb.ConfigCenterAddRes{}, nil
}

func (c ctrl) ConfigCenterList(ctx context.Context, req *adminpb.ConfigCenterListReq) (*adminpb.ConfigCenterListRes, error) {
	caller := auth.ExtractAdminUser(ctx).Uid
	list, total, err := dao.AdminConfigCenterDao.List(ctx, req)
	if err != nil {
		return nil, err
	}
	return &adminpb.ConfigCenterListRes{
		List: lo.Map(list, func(item *admin.AdminConfigCenter, _ int) *commonpb.ConfigItem {
			pb := item.ToPB()
			if item.IsLock {
				if !(caller != 0 && caller == item.CreatedBy) {
					pb.Core.Value = "【仅创建人可见&可改】"
				}
			}
			return pb
		}),
		Total: total,
	}, nil
}

func (c ctrl) ConfigCenterUpdate(ctx context.Context, req *adminpb.ConfigCenterUpdateReq) (*adminpb.ConfigCenterUpdateRes, error) {
	if req.Item == nil {
		return nil, xerr.ErrParams.New("请提供item")
	}
	if req.Item.Key == "" {
		return nil, xerr.ErrParams.New("请提供key")
	}
	caller := auth.ExtractAdminUser(ctx)
	err := dao.AdminConfigCenterDao.UpdateConfigValue(ctx, req.Item, false, caller)
	return &adminpb.ConfigCenterUpdateRes{}, err
}

func (c ctrl) ConfigCenterDelete(ctx context.Context, req *adminpb.ConfigCenterDeleteReq) (*adminpb.ConfigCenterDeleteRes, error) {
	caller := auth.ExtractAdminUser(ctx)
	err := dao.AdminConfigCenterDao.Delete(ctx, req.Key, caller)
	return &adminpb.ConfigCenterDeleteRes{}, err
}
