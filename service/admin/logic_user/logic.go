package logic_user

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/infra/svccli/rpc"
	"microsvc/model/svc/micro_svc"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/commonpb"
	"microsvc/service/admin/dao"

	"github.com/samber/lo"
)

type ctrl struct {
}

var Ext ctrl

func (ctrl) ListBizUser(ctx context.Context, req *adminpb.ListUserReq) (*adminpb.ListUserRes, error) {
	list, total, err := dao.BizUser.ListUser(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return &adminpb.ListUserRes{Total: total}, nil
	}
	uids := lo.Map(list, func(item *user.User, _ int) int64 { return item.Uid })

	// 获取金币相关
	goldMap, err := __getGold(ctx, uids)
	if err != nil {
		return nil, err
	}
	// 获取每个用户的最后登录日志
	signInMap, err := __getSignInLog(ctx, uids)
	if err != nil {
		return nil, err
	}
	// 获取每个用户的生效中的惩罚记录
	punishMap, err := __getPunish(ctx, uids)
	if err != nil {
		return nil, err
	}
	res := &adminpb.ListUserRes{
		List: lo.Map(list, func(item *user.User, _ int) *adminpb.UserInfo {
			return &adminpb.UserInfo{
				User:          item.ToPB(),
				Gold:          goldMap[item.Uid],
				LastSignInLog: signInMap[item.Uid],
				Punish:        punishMap[item.Uid],
				Terminate:     &commonpb.UserTerminate{},
			}
		}),
		Total: total,
	}
	return res, nil
}

func (ctrl) ListUserAPICallLog(ctx context.Context, req *adminpb.ListUserAPICallLogReq) (*adminpb.ListUserAPICallLogRes, error) {
	list, total, err := dao.BizUser.ListUserAPICallLog(ctx, req)
	return &adminpb.ListUserAPICallLogRes{
		List: lo.Map(list, func(item *micro_svc.APICallLog, _ int) *adminpb.APICallLog {
			return item.ToPB()
		}),
		Total: total,
	}, err
}

func (c ctrl) ListUserLastSignInLogs(ctx context.Context, req *adminpb.ListUserLastSignInLogsReq) (*adminpb.ListUserLastSignInLogsRes, error) {
	list, err := dao.BizUser.GetLastSignInLogs(ctx, req.Uid, req.Limit)
	return &adminpb.ListUserLastSignInLogsRes{
		List: lo.Map(list, func(item *user.SignInLog, _ int) *commonpb.UserSignInLog {
			return item.ToPB()
		}),
	}, err
}

func (ctrl) NewPunish(ctx context.Context, req *adminpb.NewPunishReq) (*adminpb.NewPunishRes, error) {
	if req.Inner == nil {
		return nil, xerr.ErrParams.New("`inner`字段不能为空")
	}
	req.Inner.AdminUid = auth.GetAuthUID(ctx)
	res, err := rpc.User().NewPunish(ctx, req.Inner)
	return &adminpb.NewPunishRes{
		Inner: res,
	}, err
}

func (ctrl) IncrPunishDuration(ctx context.Context, req *adminpb.IncrPunishDurationReq) (*adminpb.IncrPunishDurationRes, error) {
	if req.Inner == nil {
		return nil, xerr.ErrParams.New("`inner`字段不能为空")
	}
	req.Inner.AdminUid = auth.GetAuthUID(ctx)
	res, err := rpc.User().IncrPunishDuration(ctx, req.Inner)
	return &adminpb.IncrPunishDurationRes{
		Inner: res,
	}, err
}

func (ctrl) DismissPunish(ctx context.Context, req *adminpb.DismissPunishReq) (*adminpb.DismissPunishRes, error) {
	if req.Inner == nil {
		return nil, xerr.ErrParams.New("`inner`字段不能为空")
	}
	req.Inner.AdminUid = auth.GetAuthUID(ctx)
	res, err := rpc.User().DismissPunish(ctx, req.Inner)
	return &adminpb.DismissPunishRes{
		Inner: res,
	}, err
}

func (ctrl) PunishList(ctx context.Context, req *adminpb.PunishListReq) (*adminpb.PunishListRes, error) {
	if req.Inner == nil {
		return nil, xerr.ErrParams.New("`inner`字段不能为空")
	}
	res, err := rpc.User().PunishList(ctx, req.Inner)
	return &adminpb.PunishListRes{
		Inner: res,
	}, err
}

func (ctrl) UserPunishLog(ctx context.Context, req *adminpb.UserPunishLogReq) (*adminpb.UserPunishLogRes, error) {
	if req.Inner == nil {
		return nil, xerr.ErrParams.New("`inner`字段不能为空")
	}
	res, err := rpc.User().UserPunishLog(ctx, req.Inner)
	return &adminpb.UserPunishLogRes{
		Inner: res,
	}, err
}
