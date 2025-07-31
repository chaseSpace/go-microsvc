package handler

import (
	"context"
	"microsvc/protocol/svc/crontaskpb"
	"microsvc/service/crontask/logic"
)

var IntCtrl crontaskpb.CrontaskIntServer = new(intCtrl)

type intCtrl struct{}

func (t intCtrl) TestInt(ctx context.Context, req *crontaskpb.TestIntReq) (*crontaskpb.TestIntRes, error) {
	return logic.Int.TestInt(ctx, req)
}
