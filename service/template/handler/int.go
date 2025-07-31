package handler

import (
	"context"
	"microsvc/protocol/svc/templatepb"
	"microsvc/service/template/logic"
)

var IntCtrl templatepb.TemplateIntServer = new(intCtrl)

type intCtrl struct{}

func (t intCtrl) TestInt(ctx context.Context, req *templatepb.TestIntReq) (*templatepb.TestIntRes, error) {
	return logic.Int.TestInt(ctx, req)
}
