package logic_user

import (
	"context"
	"microsvc/infra/svccli/rpc"
	"microsvc/model/svc/currency"
	"microsvc/model/svc/user"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/admin/dao"

	"github.com/samber/lo"
)

func __getGold(ctx context.Context, uids []int64) (map[int64]*commonpb.UserGold, error) {
	list, err := dao.BizGold.GetMultiUserAccount(ctx, uids)
	if err != nil {
		return nil, err
	}
	vmap := lo.SliceToMap(list, func(item *currency.GoldAccount) (int64, *commonpb.UserGold) {
		return item.UID, &commonpb.UserGold{
			Balance:       item.Balance,
			RechargeTotal: item.RechargeTotal}
	})
	// 填零值
	for _, uid := range uids {
		if vmap[uid] == nil {
			vmap[uid] = &commonpb.UserGold{}
		}
	}
	// todo 获取消费总额
	return vmap, nil
}

func __getSignInLog(ctx context.Context, uids []int64) (map[int64]*commonpb.UserSignInLog, error) {
	list, err := dao.BizUser.GetLatestSignInLog(ctx, uids)
	if err != nil {
		return nil, err
	}
	vmap := lo.SliceToMap(list, func(item *user.SignInLog) (int64, *commonpb.UserSignInLog) {
		return item.UID, item.ToPB()
	})
	// 填零值
	for _, uid := range uids {
		if vmap[uid] == nil {
			vmap[uid] = &commonpb.UserSignInLog{}
		}
	}
	return vmap, nil
}

func __getPunish(ctx context.Context, uids []int64) (vmap map[int64][]*commonpb.UserPunish, err error) {
	res, err := rpc.User().PunishList(ctx, &userpb.PunishListReq{
		SearchUid:   uids,
		SearchState: commonpb.PunishState_PS_InProgress,
		Page: &commonpb.PageArgs{
			Pn: 1,
			Ps: 10,
			//IsDownload: true,
		},
	})
	if err != nil {
		return
	}
	if res.Total == 0 {
		return
	}
	vmap = make(map[int64][]*commonpb.UserPunish)
	for _, v := range res.List {
		vmap[v.Uid] = append(vmap[v.Uid], &commonpb.UserPunish{
			Type:      v.Type,
			Duration:  v.Duration,
			Reason:    v.Reason,
			State:     v.State,
			CreatedAt: v.CreatedAt,
		})
	}
	return
}
