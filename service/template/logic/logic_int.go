package logic

import (
	"context"
	"microsvc/protocol/svc/templatepb"
)

type intCtrl struct{}

var Int = intCtrl{} // 暴露struct而不是interface，方便IDE跳转

func (intCtrl) TestInt(ctx context.Context, req *templatepb.TestIntReq) (*templatepb.TestIntRes, error) {
	panic("implement me")
}
