package logic

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/consts"
	"microsvc/model/svc/friend"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/friendpb"
	"microsvc/service/friend/cache"
	"microsvc/service/friend/dao"
	"strings"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

type ctrl struct {
}

var Ext ctrl

func (ctrl) FriendList(ctx context.Context, caller *auth.SvcCaller, req *friendpb.FriendListReq) (*friendpb.FriendListRes, error) {
	list, total, err := cache.FriendCtrl.FriendList(ctx, caller.Uid, req)
	if err != nil {
		return nil, err
	}

	friendItems := lo.Map(list, func(item *friend.FriendRPC, index int) *friendpb.Friend {
		return item.ToPB()
	})
	return &friendpb.FriendListRes{
		List:  friendItems,
		Total: total,
	}, nil
}

func (ctrl) FriendOnewayList(ctx context.Context, caller *auth.SvcCaller, req *friendpb.FriendOnewayListReq) (*friendpb.FriendOnewayListRes, error) {
	list, total, err := cache.FriendCtrl.OnewayList(ctx, caller.Uid, req)
	if err != nil {
		return nil, err
	}

	friendItems := lo.Map(list, func(item *friend.FriendRPC, index int) *friendpb.Friend {
		return item.ToPB()
	})
	return &friendpb.FriendOnewayListRes{
		List:  friendItems,
		Total: total,
	}, nil
}

func (ctrl) FollowOne(ctx context.Context, caller *auth.SvcCaller, req *friendpb.FollowOneReq) (res *friendpb.FollowOneRes, err error) {
	tx := friend.Q.WithContext(ctx).Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	count, err := dao.GetFriendCnt(ctx, tx, caller.Uid)
	if err != nil {
		return
	}
	if count >= consts.MaxFriendCount {
		return nil, xerr.ErrFriendCountUpToMax
	}
	err = dao.FollowOne(ctx, tx, caller.Uid, req.TargetUid)
	if err != nil {
		return nil, err
	}

	// 查询现在的单边关系
	var onewayRows []*friend.Friend
	onewayRows, err = dao.GetOnewayData(ctx, tx, caller.Uid, req.TargetUid)
	if err != nil {
		return
	}

	// 删除双方单边关系缓存
	err = cache.FriendCtrl.CleanOnewayList(ctx, caller.Uid, req.TargetUid, true)
	if err != nil {
		return
	}

	res = &friendpb.FollowOneRes{}

	// 互关时，删除双方好友列表缓存
	if len(onewayRows) == 2 {
		err = cache.FriendCtrl.CleanFriendList(ctx, caller.Uid, req.TargetUid)
		if err != nil {
			return
		}
		res.Mutual = true
	}
	return
}

func (ctrl) __unFollowOneCore(ctx context.Context, tx *gorm.DB, uid, targetUID int64) (err error) {
	err = dao.UnFollowOne(ctx, tx, uid, targetUID)
	if err != nil {
		return
	}
	// 删除双方单边关系缓存
	err = cache.FriendCtrl.CleanOnewayList(ctx, uid, targetUID, true)
	return
}

func (c ctrl) UnFollowOne(ctx context.Context, caller *auth.SvcCaller, req *friendpb.FollowOneReq) (res *friendpb.FollowOneRes, err error) {
	tx := friend.Q.WithContext(ctx).Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	res = &friendpb.FollowOneRes{}

	err = c.__unFollowOneCore(ctx, tx, caller.Uid, req.TargetUid)
	if err != nil {
		return
	}

	// 查询互关数据
	var onewayRows []*friend.Friend
	onewayRows, err = dao.GetOnewayData(ctx, tx, caller.Uid, req.TargetUid)
	if err != nil {
		return
	}
	// 互关时，删除双方好友列表缓存
	if len(onewayRows) == 2 {
		err = cache.FriendCtrl.CleanFriendList(ctx, caller.Uid, req.TargetUid)
	}
	return
}

func (ctrl) SearchFriendList(ctx context.Context, caller *auth.SvcCaller, req *friendpb.SearchFriendListReq) (res *friendpb.SearchFriendListRes, err error) {
	var listDao []*friend.Friend
	var list []*friend.FriendRPC

	req.Keyword = strings.TrimSpace(req.Keyword)

	// 接口限速后，SQL直查的方式够用很长一段时间，实测百万级记录搜索耗时控制100ms内（16g MAC）
	// 后期可采用专门的搜索引擎提高全文搜索效率
	listDao, err = dao.SearchFriendList(ctx, caller.Uid, req)
	if err != nil {
		return nil, err
	}

	// 用户信息每次都要重新获取（有单独的用户信息缓存）
	list, err = cache.SetupFriendUserInfo(ctx, listDao)
	if err != nil {
		return
	}
	res = &friendpb.SearchFriendListRes{
		List: lo.Map(list, func(item *friend.FriendRPC, index int) *friendpb.Friend {
			return item.ToPB()
		}),
	}
	return res, nil
}

func (ctrl) SearchFriendOnewayList(ctx context.Context, caller *auth.SvcCaller, req *friendpb.SearchFriendOnewayListReq) (res *friendpb.SearchFriendOnewayListRes, err error) {
	var listDao []*friend.Friend
	var list []*friend.FriendRPC

	req.Keyword = strings.TrimSpace(req.Keyword)

	// 接口限速后，SQL直查的方式够用很长一段时间，实测百万级记录搜索耗时控制100ms内（16g MAC）
	// 后期可采用专门的搜索引擎提高全文搜索效率
	listDao, err = dao.SearchFriendOnewayList(ctx, caller.Uid, req, 20)
	if err != nil {
		return nil, err
	}

	// 用户信息每次都要重新获取（有单独的用户信息缓存）
	list, err = cache.SetupFriendUserInfo(ctx, listDao)
	if err != nil {
		return
	}
	res = &friendpb.SearchFriendOnewayListRes{
		List: lo.Map(list, func(item *friend.FriendRPC, index int) *friendpb.Friend {
			return item.ToPB()
		}),
	}
	return res, nil
}

func (c ctrl) BlockOne(ctx context.Context, caller *auth.SvcCaller, req *friendpb.BlockOneReq) (res *friendpb.BlockOneRes, err error) {
	res = &friendpb.BlockOneRes{}
	if !req.IsBlock {
		err = dao.UnBlockOne(ctx, caller.Uid, req.TargetUid)
		return res, err
	}
	tx := friend.Q.WithContext(ctx).Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = dao.BlockOne(ctx, tx, caller.Uid, req.TargetUid)
	if err != nil {
		return nil, err
	}

	// 查询互关数据
	var onewayRows []*friend.Friend
	onewayRows, err = dao.GetOnewayData(ctx, tx, caller.Uid, req.TargetUid)
	if err != nil {
		return
	}
	if len(onewayRows) == 0 { // 没有关系
		return
	}

	isFollow := false
	isFans := false
	if len(onewayRows) == 2 {
		isFollow = true
		isFans = true
	} else if len(onewayRows) == 1 {
		if onewayRows[0].UID == caller.Uid {
			isFollow = true
		} else {
			isFans = true
		}
	}

	if isFollow {
		// 拉黑后，将取消双方的关注关系（含好友）
		err = c.__unFollowOneCore(ctx, tx, caller.Uid, req.TargetUid)
		if err != nil {
			return nil, err
		}
	}

	if isFans {
		err = c.__unFollowOneCore(ctx, tx, req.TargetUid, caller.Uid)
		if err != nil {
			return
		}
	}

	// 互关时，删除双方好友列表缓存
	if isFollow && isFans {
		err = cache.FriendCtrl.CleanFriendList(ctx, caller.Uid, req.TargetUid)
		if err != nil {
			return
		}
	}
	return res, err
}

// BlockList 非高频接口，不用缓存
func (ctrl) BlockList(ctx context.Context, caller *auth.SvcCaller, req *friendpb.BlockListReq) (*friendpb.BlockListRes, error) {
	list, total, err := dao.BlockList(ctx, caller.Uid, req)
	if err != nil {
		return nil, err
	}
	res := &friendpb.BlockListRes{}
	if total == 0 {
		return res, nil
	}
	listRPC, err := cache.SetupBlockUserInfo(ctx, list)
	if err != nil {
		return nil, err
	}
	return &friendpb.BlockListRes{
		Total: total,
		List: lo.Map(listRPC, func(item *friend.BlockRPC, index int) *friendpb.BlockUser {
			return item.ToPB()
		}),
	}, nil
}

func (ctrl) RelationWithOne(ctx context.Context, caller *auth.SvcCaller, req *friendpb.RelationWithOneReq) (*friendpb.RelationWithOneRes, error) {
	res := &friendpb.RelationWithOneRes{}

	// 1. 关注/粉丝/好友
	oneway, err := dao.GetOnewayData(ctx, nil, caller.Uid, req.TargetUid)
	if err != nil {
		return nil, err
	}
	if len(oneway) == 2 {
		res.Relation = friendpb.RelationType_RT_Friend
		return res, nil
	} else if len(oneway) == 1 {
		if oneway[0].UID == caller.Uid {
			res.Relation = friendpb.RelationType_RT_Follow
		} else {
			res.Relation = friendpb.RelationType_RT_Fans
		}
		return res, nil
	}

	// 2. 拉黑关系
	blocks, err := dao.GetBlockData(ctx, nil, caller.Uid, req.TargetUid)
	if err != nil {
		return nil, err
	}
	if len(blocks) == 2 {
		res.Relation = friendpb.RelationType_RT_MutualBlock
		return res, nil
	} else if len(blocks) == 1 {
		if blocks[0].UID == caller.Uid {
			res.Relation = friendpb.RelationType_RT_Block
		} else {
			res.Relation = friendpb.RelationType_RT_BeBlock
		}
	}
	return res, nil
}

func (ctrl) SaveVisitor(ctx context.Context, caller *auth.SvcCaller, req *friendpb.SaveVisitorReq) (*friendpb.SaveVisitorRes, error) {
	allow, err := cache.VisitorCtrl.AllowSaveVisitor(ctx, caller.Uid, req.TargetUid, &req.Seconds)
	if err != nil {
		return nil, err
	}
	if !allow {
		return nil, xerr.ErrParams.New("保存访客：前端计时有误")
	}
	err = dao.SaveVisitor(ctx, req.TargetUid, caller.Uid, req.Seconds)
	if err != nil {
		return nil, err
	}
	// 删除目标用户的访客列表缓存
	err = cache.VisitorCtrl.CleanVisitorList(ctx, req.TargetUid)
	res := &friendpb.SaveVisitorRes{}
	return res, err
}

func (ctrl) VisitorList(ctx context.Context, caller *auth.SvcCaller, req *friendpb.VisitorListReq) (*friendpb.VisitorListRes, error) {
	listRPC, total, visitorsTotal, visitorsRepeated, err := cache.VisitorCtrl.VisitorList(ctx, caller.Uid, req)
	if err != nil {
		return nil, err
	}
	res := &friendpb.VisitorListRes{}
	res.Total = total
	res.VisitorsTotal = visitorsTotal
	res.VisitorsRepeated = visitorsRepeated
	res.List = lo.Map(listRPC, func(item *friend.VisitorRPC, index int) *friendpb.Visitor {
		return item.ToPB()
	})
	return res, nil
}
