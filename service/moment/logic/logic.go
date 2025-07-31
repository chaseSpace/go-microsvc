package logic

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/bizcomm/commrpc"
	"microsvc/model/svc/moment"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/momentpb"
	"microsvc/service/moment/cache"
	"microsvc/service/moment/dao"
	"unicode/utf8"

	"github.com/samber/lo"
)

type ctrl struct {
}

var Ext ctrl

func (c ctrl) CreateMoment(ctx context.Context, uid int64, req *momentpb.CreateMomentReq) (*momentpb.CreateMomentRes, error) {
	err := __checkNewMoment(ctx, req)
	if err != nil {
		return nil, err
	}
	// 机审
	result, ReqId, err := __execAIReview(ctx, uid, req)
	if err != nil {
		return nil, err
	}
	// AI审核不过的时候，直接返回
	if result.GetStatus() == commonpb.AIReviewStatus_ARS_Reject {
		return nil, xerr.ErrReviewResultReject
	}
	// 写入动态表
	mid, err := dao.MomentDao.CreateMoment(ctx, uid, req)
	if err != nil {
		return nil, err
	}
	// 报送管理后台审核
	err = __AddAdminReview(ctx, uid, req, mid, ReqId, result.GetStatus())
	if err != nil {
		return nil, err
	}
	return &momentpb.CreateMomentRes{
		Mid:           mid,
		WaitingReview: true,
	}, nil
}

func (c ctrl) DeleteMoment(ctx context.Context, uid int64, req *momentpb.DeleteMomentReq) (*momentpb.DeleteMomentRes, error) {
	err := dao.MomentDao.DeleteMoment(ctx, uid, req.Mid)
	return &momentpb.DeleteMomentRes{}, err
}

func (c ctrl) LikeMoment(ctx context.Context, caller int64, req *momentpb.LikeMomentReq) (*momentpb.LikeMomentRes, error) {
	err := dao.MomentDao.LikeMoment(ctx, caller, req.IsLike, req.Mid)
	return &momentpb.LikeMomentRes{}, err
}

func (c ctrl) CommentMoment(ctx context.Context, caller int64, req *momentpb.CommentMomentReq) (*momentpb.CommentMomentRes, error) {
	if utf8.RuneCountInString(req.Content) > 100 {
		return nil, xerr.ErrCommentTextTooLong
	}
	_, err := dao.MomentDao.GetMoment(ctx, req.Mid, true)
	if err != nil {
		return nil, err
	}
	err = dao.MomentDao.CommentMoment(ctx, caller, req.Mid, req.ReplyUid, req.Content)
	return &momentpb.CommentMomentRes{}, err
}

func (c ctrl) ForwardMoment(ctx context.Context, caller int64, req *momentpb.ForwardMomentReq) (res *momentpb.ForwardMomentRes, err error) {
	var yes bool // 通过缓存几天转发操作来避免同个人操作频繁增加转发数
	if yes, err = cache.MomentCache.NeverForward(ctx, req.Mid, caller); err != nil {
		return nil, err
	} else if yes {
		err = dao.MomentDao.ForwardMoment(ctx, req.Mid)
	}
	return &momentpb.ForwardMomentRes{}, err
}

func (c ctrl) GetComment(ctx context.Context, req *momentpb.GetCommentReq) (*momentpb.GetCommentRes, error) {
	list, total, err := dao.MomentDao.GetComment(ctx, req)
	if err != nil {
		return nil, err
	}
	res := &momentpb.GetCommentRes{}
	if total == 0 {
		return res, nil
	}

	var uids []int64
	for _, item := range list {
		uids = append(uids, item.UID)
		if item.ReplyUID != 0 {
			uids = append(uids, item.ReplyUID)
		}
	}

	// 填充用户信息
	err = commrpc.PopulateUserBase(ctx, list)
	if err != nil {
		return nil, err
	}
	res.Total = total
	res.List = lo.Map(list, func(item *moment.MomentComment, _ int) *momentpb.CommentMix {
		return &momentpb.CommentMix{
			Comment:   item.ToPB(),
			User:      item.User,
			ReplyUser: item.ReplyUser,
		}
	})
	return res, nil
}

func (c ctrl) ListFollowMoment(ctx context.Context, uid int64, req *momentpb.ListFollowMomentReq) (*momentpb.ListFollowMomentRes, error) {
	res := &momentpb.ListFollowMomentRes{LastIndex: -1}
	if req.LastIndex < 0 {
		return res, nil
	}
	list, err := dao.MomentDao.ListFollowMoment(ctx, uid, req.LastIndex, req.PageSize+1)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return res, nil
	}
	if len(list) == int(req.PageSize) {
		list = list[:req.PageSize-1]
		res.LastIndex = list[len(list)-1].ReviewPassAt
	}

	// 填充用户信息
	err = commrpc.PopulateUserBase(ctx, list)
	if err != nil {
		return nil, err
	}

	mids := lo.Map(list, func(item *moment.Moment, _ int) int64 { return item.Id })
	// 获取评论数
	commentsMap, err := dao.MomentDao.GetMomentCommentNum(ctx, mids)
	if err != nil {
		return nil, err
	}
	res.List = lo.Map(list, func(item *moment.Moment, _ int) *momentpb.MomentMix {
		return &momentpb.MomentMix{
			Moment: item.ToPB(commentsMap[item.Id]),
			User:   item.User,
		}
	})
	return res, nil
}

func (c ctrl) ListLatestMoment(ctx context.Context, caller *auth.SvcCaller, req *momentpb.ListLatestMomentReq) (*momentpb.ListLatestMomentRes, error) {
	list, err := dao.MomentDao.ListLatestMoment(ctx, req.LastIndex, req.PageSize+1, caller.Sex)
	if err != nil {
		return nil, err
	}
	res := &momentpb.ListLatestMomentRes{LastIndex: -1}
	if len(list) == 0 {
		return res, nil
	}
	if len(list) == int(req.PageSize) {
		list = list[:req.PageSize-1]
		res.LastIndex = list[len(list)-1].ReviewPassAt
	}

	// 填充用户信息
	err = commrpc.PopulateUserBase(ctx, list)
	if err != nil {
		return nil, err
	}

	// 获取评论数
	mids := lo.Map(list, func(item *moment.Moment, _ int) int64 { return item.Id })
	commentsMap, err := dao.MomentDao.GetMomentCommentNum(ctx, mids)
	if err != nil {
		return nil, err
	}

	res.List = lo.Map(list, func(item *moment.Moment, _ int) *momentpb.MomentMix {
		return &momentpb.MomentMix{
			Moment: item.ToPB(commentsMap[item.Id]),
			User:   item.User,
		}
	})
	return res, nil
}
