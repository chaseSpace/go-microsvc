package logic_review

import (
	"context"
	"microsvc/protocol/svc/thirdpartypb"
)

type ctrl struct {
}

var Ext ctrl

func (c ctrl) SyncReviewImageExt(ctx context.Context, caller int64, req *thirdpartypb.SyncReviewImageExtReq) (*thirdpartypb.SyncReviewImageExtRes, error) {
	rr, err := provider.ReviewImage(ctx, caller, req.Uri, req.Type, req.Ext)
	if err != nil {
		return nil, err
	}
	res := &thirdpartypb.SyncReviewImageExtRes{Status: rr.Status, Message: rr.RiskDescription}
	return res, nil
}
