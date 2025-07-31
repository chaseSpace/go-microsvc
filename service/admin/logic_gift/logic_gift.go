package logic_gift

import (
	"context"
	"microsvc/infra/svccli/rpc"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/adminpb"
)

type ctrl struct {
}

var Ext ctrl

func (ctrl) GetGiftList(ctx context.Context, req *adminpb.GetGiftListReq) (*adminpb.GetGiftListRes, error) {
	if req.Inner == nil {
		return nil, xerr.ErrParams.New("Field `inner` is required")
	}
	res, err := rpc.Gift().GetGiftListInt(ctx, req.Inner)
	return &adminpb.GetGiftListRes{
		Inner: res,
	}, err
}

func (ctrl) SaveGiftItem(ctx context.Context, req *adminpb.SaveGiftItemReq) (*adminpb.SaveGiftItemRes, error) {
	if req.Inner == nil {
		return nil, xerr.ErrParams.New("Field `inner` is required")
	}
	res, err := rpc.Gift().SaveGiftItem(ctx, req.Inner)
	return &adminpb.SaveGiftItemRes{
		Inner: res,
	}, err
}

func (ctrl) DelGiftItem(ctx context.Context, req *adminpb.DelGiftItemReq) (*adminpb.DelGiftItemRes, error) {
	if req.Inner == nil {
		return nil, xerr.ErrParams.New("Field `inner` is required")
	}
	res, err := rpc.Gift().DelGiftItem(ctx, req.Inner)
	return &adminpb.DelGiftItemRes{
		Inner: res,
	}, err
}

func (ctrl) GetUserGiftTxLog(ctx context.Context, req *adminpb.GetUserGiftTxLogReq) (*adminpb.GetUserGiftTxLogRes, error) {
	if req.Inner == nil {
		return nil, xerr.ErrParams.New("Field `inner` is required")
	}
	res, err := rpc.Gift().GetUserGiftTxLog(ctx, req.Inner)
	return &adminpb.GetUserGiftTxLogRes{
		Inner: res,
	}, err
}
