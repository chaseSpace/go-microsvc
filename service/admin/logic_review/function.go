package logic_review

import (
	"context"
	"microsvc/infra/svccli/rpc"
	"microsvc/model/svc/admin"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/momentpb"
	"microsvc/protocol/svc/userpb"
	"microsvc/service/admin/dao"

	"github.com/samber/lo"
)

func __getAdministratorInfo(ctx context.Context, uids []int64) (map[int64]*admin.AdminUser, error) {
	vmap := make(map[int64]*admin.AdminUser)
	if len(uids) > 0 {
		models, _, err := dao.GetUser(ctx, uids...)
		if err != nil {
			return nil, err
		}
		lo.ForEach(models, func(item *admin.AdminUser, index int) {
			vmap[item.Uid] = item
		})
		for _, uid := range uids {
			if vmap[uid] == nil {
				vmap[uid] = &admin.AdminUser{
					Uid:      uid,
					Nickname: "未知管理员",
				}
			}
		}
	}
	return vmap, nil
}

type reviewItem interface {
	GetUpdateByAdminUID() int64
	GetUID() int64
}

func __getPeopleInfoMap[T reviewItem](ctx context.Context, entities []T) (adminMap map[int64]*admin.AdminUser, userMap map[int64]*commonpb.User, err error) {
	// 查询管理员信息
	uids := lo.Map(entities, func(item T, _ int) int64 { return item.GetUpdateByAdminUID() })
	uids = lo.Filter(uids, func(item int64, _ int) bool { return item != 0 })
	adminMap, err = __getAdministratorInfo(ctx, uids)
	if err != nil {
		return
	}

	// 查询用户信息
	uids2 := lo.Map(entities, func(item T, _ int) int64 { return item.GetUID() })
	var res *userpb.GetUserInfoIntRes
	res, err = rpc.User().GetUserInfoInt(ctx, &userpb.GetUserInfoIntReq{Uids: uids2, PopulateNotfound: true})
	if err != nil {
		return
	}
	userMap = res.Umap
	return
}

func __tryUpdateReviewText(ctx context.Context, caller int64, req *adminpb.UpdateReviewStatusReq) error {
	list, _, err := dao.ReviewDao.ListText(ctx, &adminpb.ListReviewTextReq{Id: req.Id})
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return xerr.ErrDataNotExist
	}
	data := list[0] // 任何状态都可以 再次更新
	hits, err := dao.ReviewDao.UpdateText(ctx, data.Id, data.UID, caller, data.Status, req.Status)
	if err != nil {
		return err
	}
	if !hits {
		return xerr.ErrNoRowAffectedOnUpdate
	}

	isPass := req.Status == commonpb.ReviewStatus_RS_ManualPass // 此时只有 通过/拒绝
	// 通知对应服务
	switch data.BizType {
	case commonpb.BizType_RBT_Nickname, commonpb.BizType_RBT_UserDesc:
		_, err = rpc.User().ReviewProfile(ctx, &userpb.ReviewProfileReq{Uid: data.UID, IsPass: isPass, Reason: req.Note, BizType: data.BizType})
	}
	return err
}

func __tryUpdateReviewImage(ctx context.Context, caller int64, req *adminpb.UpdateReviewStatusReq) error {
	list, _, err := dao.ReviewDao.ListImage(ctx, &adminpb.ListReviewImageReq{Id: req.Id})
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return xerr.ErrDataNotExist
	}
	data := list[0] // 任何状态都可以 再次更新
	hits, err := dao.ReviewDao.UpdateImage(ctx, data.Id, data.UID, caller, data.Status, req.Status)
	if err != nil {
		return err
	}
	if !hits {
		return xerr.ErrNoRowAffectedOnUpdate
	}

	isPass := req.Status == commonpb.ReviewStatus_RS_ManualPass // 此时只有 通过/拒绝
	// 通知对应服务
	switch data.BizType {
	case commonpb.BizType_RBT_Album, commonpb.BizType_RBT_Avatar:
		_, err = rpc.User().ReviewProfile(ctx, &userpb.ReviewProfileReq{Uid: data.UID, IsPass: isPass, Reason: req.Note, BizType: data.BizType})

	}
	return err
}

func __tryUpdateReviewVideo(ctx context.Context, caller int64, req *adminpb.UpdateReviewStatusReq) error {
	list, _, err := dao.ReviewDao.ListVideo(ctx, &adminpb.ListReviewVideoReq{Id: req.Id})
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return xerr.ErrDataNotExist
	}
	data := list[0] // 任何状态都可以 再次更新
	hits, err := dao.ReviewDao.UpdateVideo(ctx, data.Id, data.UID, caller, data.Status, req.Status)
	if err != nil {
		return err
	}
	if !hits {
		return xerr.ErrNoRowAffectedOnUpdate
	}

	// 通知对应服务
	switch data.BizType {
	case commonpb.BizType_RBT_Moment:
		momentStatus := map[commonpb.ReviewStatus]momentpb.ReviewStatus{
			commonpb.ReviewStatus_RS_ManualPass:   momentpb.ReviewStatus_RS_Pass,
			commonpb.ReviewStatus_RS_ManualReject: momentpb.ReviewStatus_RS_Reject,
		}[req.Status]
		_, err = rpc.Moment().UpdateReviewStatus(ctx, &momentpb.UpdateReviewStatusReq{Uid: data.UID, Mid: data.BizUniqId, Status: momentStatus})
	}
	return err
}
