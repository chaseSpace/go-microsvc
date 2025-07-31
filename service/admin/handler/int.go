package handler

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/model/svc/admin"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/commonpb"
	"microsvc/service/admin/dao"
	"microsvc/service/admin/logic_review"

	"github.com/samber/lo"
)

type intCtrl struct{}

var IntCtrl adminpb.AdminIntServer = new(intCtrl)

func (a intCtrl) AddReview(ctx context.Context, req *adminpb.AddReviewReq) (*adminpb.AddReviewRes, error) {
	return logic_review.Int.AddReview(ctx, req)
}

func (a intCtrl) ConfigCenterGet(ctx context.Context, req *adminpb.ConfigCenterGetReq) (*adminpb.ConfigCenterGetRes, error) {
	list, err := dao.AdminConfigCenterDao.GetConfigByKeys(ctx, req.Keys...)
	if err != nil {
		return nil, err
	}
	return &adminpb.ConfigCenterGetRes{
		Cmap: lo.SliceToMap(list, func(item *admin.AdminConfigCenter) (string, *commonpb.ConfigItem) {
			return item.Key, item.ToPB()
		}),
	}, nil
}

func (a intCtrl) ConfigCenterUpdateInt(ctx context.Context, req *adminpb.ConfigCenterUpdateIntReq) (*adminpb.ConfigCenterUpdateIntRes, error) {
	err := dao.AdminConfigCenterDao.UpdateConfigValue(ctx, req.Item, true, auth.NewFakeAdminCaller())
	if xerr.ErrDataNotExist.Equal(err) && req.AddOnNotExist {
		err = dao.AdminConfigCenterDao.Add(ctx, &commonpb.ConfigItem{
			Core:      req.Item,
			CreatedBy: 0, // 0表示程序创建
			UpdatedBy: 0,
		})
	}
	return new(adminpb.ConfigCenterUpdateIntRes), err
}

func (a intCtrl) SwitchCenterGet(ctx context.Context, req *adminpb.SwitchCenterGetReq) (*adminpb.SwitchCenterGetRes, error) {
	list, err := dao.AdminSwitchCenterDao.GetSwitchByKeys(ctx, req.Keys...)
	if err != nil {
		return nil, err
	}
	return &adminpb.SwitchCenterGetRes{
		Smap: lo.SliceToMap(list, func(item *admin.AdminSwitchCenter) (string, *commonpb.SwitchItem) {
			return item.Key, item.ToPB()
		}),
	}, nil
}
