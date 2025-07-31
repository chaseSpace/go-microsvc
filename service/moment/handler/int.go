package handler

import (
	"context"
	"microsvc/protocol/svc/momentpb"
	"microsvc/service/moment/logic"
)

var IntCtrl momentpb.MomentIntServer = new(intCtrl)

type intCtrl struct{}

func (i intCtrl) UpdateReviewStatus(ctx context.Context, req *momentpb.UpdateReviewStatusReq) (*momentpb.UpdateReviewStatusRes, error) {
	return logic.Int.UpdateReviewStatus(ctx, req)
}
