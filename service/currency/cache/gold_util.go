package cache

import (
	"context"
	"microsvc/infra/svccli/rpc"
	"microsvc/model/svc/friend"
	"microsvc/protocol/svc/userpb"

	"github.com/samber/lo"
)

func SetupFriendUserInfo(ctx context.Context, listDao []*friend.Friend) (listRPC []*friend.FriendRPC, err error) {
	if len(listDao) == 0 {
		return
	}
	res, err := rpc.User().GetUserInfoInt(ctx, &userpb.GetUserInfoIntReq{
		PopulateNotfound: true,
		Uids: lo.Map(listDao, func(v *friend.Friend, _ int) int64 {
			return v.FID
		}),
	})
	if err != nil {
		return nil, err
	}
	for _, v := range listDao {
		userPB := res.Umap[v.FID]
		if userPB == nil {
			continue
		}
		listRPC = append(listRPC, &friend.FriendRPC{
			Friend: v,
			UserPB: userPB,
		})
	}
	return
}

func SetupBlockUserInfo(ctx context.Context, listDao []*friend.Block) (listRPC []*friend.BlockRPC, err error) {
	if len(listDao) == 0 {
		return
	}
	res, err := rpc.User().GetUserInfoInt(ctx, &userpb.GetUserInfoIntReq{
		PopulateNotfound: true,
		Uids: lo.Map(listDao, func(v *friend.Block, _ int) int64 {
			return v.BID
		}),
	})
	if err != nil {
		return nil, err
	}
	for _, v := range listDao {
		userPB := res.Umap[v.BID]
		if userPB == nil {
			continue
		}
		listRPC = append(listRPC, &friend.BlockRPC{
			Block:  v,
			UserPB: userPB,
		})
	}
	return
}

func SetupVisitorUserInfo(ctx context.Context, listDao []*friend.Visitor) (listRPC []*friend.VisitorRPC, err error) {
	if len(listDao) == 0 {
		return
	}
	res, err := rpc.User().GetUserInfoInt(ctx, &userpb.GetUserInfoIntReq{
		PopulateNotfound: true,
		Uids: lo.Map(listDao, func(v *friend.Visitor, _ int) int64 {
			return v.VID
		}),
	})
	if err != nil {
		return nil, err
	}
	for _, v := range listDao {
		userPB := res.Umap[v.VID]
		if userPB == nil {
			continue
		}
		listRPC = append(listRPC, &friend.VisitorRPC{
			Visitor: v,
			UserPB:  userPB,
		})
	}
	return
}
