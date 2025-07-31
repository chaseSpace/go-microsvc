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

func (ctrl) SwitchCenterAdd(ctx context.Context, req *adminpb.SwitchCenterAddReq) (*adminpb.SwitchCenterAddRes, error) {
	caller := auth.ExtractAdminUser(ctx).Uid
	if req.Core == nil {
		return nil, xerr.ErrParams.New("请提供item")
	}
	core := req.Core
	core.Key = strings.ReplaceAll(core.Key, " ", "") // 过滤空格
	if core.Key == "" || core.Name == "" {
		return nil, xerr.ErrParams.New("请提供key/name")
	}
	if core.ValueExt == nil {
		core.ValueExt = make(map[int32]string)
	}
	err := admin.SwitchValExt(core.ValueExt).Check(core.Value)
	if err != nil {
		return nil, err
	}
	item := &commonpb.SwitchItem{
		Core:      core,
		UpdatedBy: caller,
		CreatedBy: caller,
	}
	err = dao.AdminSwitchCenterDao.Add(ctx, item)
	if xerr.ErrDataDuplicate.Equal(err) {
		return nil, errors.New("开关Key已存在")
	}
	return &adminpb.SwitchCenterAddRes{}, err
}

func (c ctrl) SwitchCenterList(ctx context.Context, req *adminpb.SwitchCenterListReq) (*adminpb.SwitchCenterListRes, error) {
	list, total, err := dao.AdminSwitchCenterDao.List(ctx, req)
	if err != nil {
		return nil, err
	}
	return &adminpb.SwitchCenterListRes{
		List: lo.Map(list, func(item *admin.AdminSwitchCenter, _ int) *commonpb.SwitchItem {
			pb := item.ToPB()
			return pb
		}),
		Total: total,
	}, nil
}

func (c ctrl) SwitchCenterUpdate(ctx context.Context, req *adminpb.SwitchCenterUpdateReq) (*adminpb.SwitchCenterUpdateRes, error) {
	caller := auth.ExtractAdminUser(ctx)
	if req.Core == nil {
		return nil, xerr.ErrParams.New("请提供item")
	}
	if req.Core.Key == "" {
		return nil, xerr.ErrParams.New("请提供skey")
	}
	err := dao.AdminSwitchCenterDao.UpdateSwitch(ctx, req.Core, caller)
	return &adminpb.SwitchCenterUpdateRes{}, err
}

func (c ctrl) SwitchCenterDelete(ctx context.Context, req *adminpb.SwitchCenterDeleteReq) (*adminpb.SwitchCenterDeleteRes, error) {
	caller := auth.ExtractAdminUser(ctx)
	err := dao.AdminSwitchCenterDao.Delete(ctx, req.Key, caller)
	return &adminpb.SwitchCenterDeleteRes{}, err
}
