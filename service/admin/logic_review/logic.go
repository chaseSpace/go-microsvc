package logic_review

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/model/svc/admin"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/commonpb"
	"microsvc/service/admin/dao"

	"github.com/samber/lo"
)

type ctrl struct {
}

var Ext ctrl

func (c ctrl) ListReviewText(ctx context.Context, req *adminpb.ListReviewTextReq) (*adminpb.ListReviewTextRes, error) {
	list, total, err := dao.ReviewDao.ListText(ctx, req)
	if err != nil || total == 0 {
		return &adminpb.ListReviewTextRes{}, err
	}

	// 查询管理员和用户信息
	adminMap, userMap, err := __getPeopleInfoMap(ctx, list)
	if err != nil {
		return nil, err
	}

	return &adminpb.ListReviewTextRes{
		List: lo.Map(list, func(item *admin.ReviewText, _ int) *adminpb.ReviewText {
			pb := item.ToPB()
			pb.AdminName = "-" // 表示暂未经过人工审核
			if model, ok := adminMap[item.UpdatedBy]; ok {
				pb.AdminName = model.Nickname
			}
			pb.User = userMap[item.UID]
			return pb
		}),
		Total: total,
	}, nil
}

func (c ctrl) ListReviewImage(ctx context.Context, req *adminpb.ListReviewImageReq) (*adminpb.ListReviewImageRes, error) {
	list, total, err := dao.ReviewDao.ListImage(ctx, req)
	if err != nil || total == 0 {
		return &adminpb.ListReviewImageRes{}, err
	}

	// 查询管理员和用户信息
	adminMap, userMap, err := __getPeopleInfoMap(ctx, list)
	if err != nil {
		return nil, err
	}

	return &adminpb.ListReviewImageRes{
		List: lo.Map(list, func(item *admin.ReviewImage, _ int) *adminpb.ReviewImage {
			pb := item.ToPB()
			pb.AdminName = "-" // 表示暂未经过人工审核
			if model, ok := adminMap[item.UpdatedBy]; ok {
				pb.AdminName = model.Nickname
			}
			pb.User = userMap[item.UID]
			return pb
		}),
		Total: total,
	}, nil
}

func (c ctrl) ListReviewVideo(ctx context.Context, req *adminpb.ListReviewVideoReq) (*adminpb.ListReviewVideoRes, error) {
	list, total, err := dao.ReviewDao.ListVideo(ctx, req)
	if err != nil || total == 0 {
		return &adminpb.ListReviewVideoRes{}, err
	}

	// 查询管理员和用户信息
	adminMap, userMap, err := __getPeopleInfoMap(ctx, list)
	if err != nil {
		return nil, err
	}

	return &adminpb.ListReviewVideoRes{
		List: lo.Map(list, func(item *admin.ReviewVideo, _ int) *adminpb.ReviewVideo {
			pb := item.ToPB()
			pb.AdminName = "-" // 表示暂未经过人工审核
			if model, ok := adminMap[item.UpdatedBy]; ok {
				pb.AdminName = model.Nickname
			}
			pb.User = userMap[item.UID]
			return pb
		}),
		Total: total,
	}, nil
}

func (c ctrl) ListReviewAudio(ctx context.Context, req *adminpb.ListReviewAudioReq) (*adminpb.ListReviewAudioRes, error) {
	list, total, err := dao.ReviewDao.ListAudio(ctx, req)
	if err != nil || total == 0 {
		return &adminpb.ListReviewAudioRes{}, err
	}

	// 查询管理员和用户信息
	adminMap, userMap, err := __getPeopleInfoMap(ctx, list)
	if err != nil {
		return nil, err
	}

	return &adminpb.ListReviewAudioRes{
		List: lo.Map(list, func(item *admin.ReviewAudio, _ int) *adminpb.ReviewAudio {
			pb := item.ToPB()
			pb.AdminName = "-" // 表示暂未经过人工审核
			if model, ok := adminMap[item.UpdatedBy]; ok {
				pb.AdminName = model.Nickname
			}
			pb.User = userMap[item.UID]
			return pb
		}),
		Total: total,
	}, nil
}

func (c ctrl) UpdateReviewStatus(ctx context.Context, req *adminpb.UpdateReviewStatusReq) (res *adminpb.UpdateReviewStatusRes, err error) {
	caller := auth.ExtractAdminUser(ctx).Uid
	switch req.Status {
	case commonpb.ReviewStatus_RS_ManualPass, commonpb.ReviewStatus_RS_ManualReject:
	default:
		return nil, xerr.ErrParams.New("不支持的status: " + req.Status.String())
	}
	switch req.Type {
	case commonpb.ReviewType_RT_Text:
		err = __tryUpdateReviewText(ctx, caller, req)
	case commonpb.ReviewType_RT_Image:
		err = __tryUpdateReviewImage(ctx, caller, req)
	case commonpb.ReviewType_RT_Video:
		err = __tryUpdateReviewVideo(ctx, caller, req)
	default:
		return nil, xerr.ErrParams.New("不支持的type:" + req.Type.String())
	}
	return &adminpb.UpdateReviewStatusRes{}, err
}
