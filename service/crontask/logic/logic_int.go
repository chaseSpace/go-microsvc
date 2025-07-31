package logic

import (
	"context"
	"microsvc/protocol/svc/crontaskpb"
)

type intCtrl struct{}

var Int = intCtrl{} // 暴露struct而不是interface，方便IDE跳转

func (intCtrl) TestInt(ctx context.Context, req *crontaskpb.TestIntReq) (*crontaskpb.TestIntRes, error) {
	panic("implement me")
}
