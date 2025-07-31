package logic_recommend

import (
	"context"
	"microsvc/protocol/svc/userpb"
)

type ctrl struct {
}

var Ext ctrl

func (c ctrl) SameCityUsers(ctx context.Context, req *userpb.SameCityUsersReq) (*userpb.SameCityUsersRes, error) {
	panic(1)
}

func (c ctrl) NearbyUsers(ctx context.Context, req *userpb.NearbyUsersReq) (*userpb.NearbyUsersRes, error) {
	//TODO implement me
	panic("implement me")
}

func (c ctrl) GetRecommendUserDetail(ctx context.Context, req *userpb.GetRecommendUserDetailReq) (*userpb.GetRecommendUserDetailRes, error) {
	//TODO implement me
	panic("implement me")
}
