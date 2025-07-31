package logic_review

import (
	"context"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/service/thirdparty/deploy"
	"microsvc/service/thirdparty/logic_review/thirdparty"
	shumeiReview "microsvc/service/thirdparty/logic_review/thirdparty/shumei"
)

var provider thirdparty.ReviewAPI

func MustInit(cc *deploy.SvcConfig) {
	provider = &shumeiReview.Shumei{}
	provider.MustInit(cc.Review.Shumei)
}

type intCtrl struct{}

var Int = intCtrl{}

func (intCtrl) SyncReviewText(ctx context.Context, req *thirdpartypb.SyncReviewTextReq) (*thirdpartypb.SyncReviewTextRes, error) {
	rr, err := provider.ReviewText(ctx, req.Uid, req.Text, req.Type, req.Ext)
	if err != nil {
		return nil, err
	}
	res := &thirdpartypb.SyncReviewTextRes{Status: rr.Status, Message: rr.RiskDescription}
	return res, nil
}

func (intCtrl) SyncReviewImage(ctx context.Context, req *thirdpartypb.SyncReviewImageReq) (*thirdpartypb.SyncReviewImageRes, error) {
	rr, err := provider.ReviewImage(ctx, req.Uid, req.Uri, req.Type, req.Ext)
	if err != nil {
		return nil, err
	}
	res := &thirdpartypb.SyncReviewImageRes{Status: rr.Status, Message: rr.RiskDescription}
	return res, nil
}

// todo 接通音频审核

func (c intCtrl) AsyncReviewVideo(ctx context.Context, req *thirdpartypb.AsyncReviewVideoReq) (*thirdpartypb.AsyncReviewVideoRes, error) {
	err := __checkReviewExtUniqReqId(req.Ext)
	if err != nil {
		return nil, err
	}
	rr, err := provider.AsyncReviewVideo(ctx, req.Uid, req.Uri, req.Type, req.Ext)
	if err != nil {
		return nil, err
	}
	res := &thirdpartypb.AsyncReviewVideoRes{ReqId: rr.ReqId, ThName: rr.ThirdPartySvcName}
	return res, nil
}

func (c intCtrl) QueryAudioReviewResult(ctx context.Context, req *thirdpartypb.QueryAudioReviewResultReq) (*thirdpartypb.QueryAudioReviewResultRes, error) {
	err := __checkReviewExtUniqReqId(req.Ext)
	if err != nil {
		return nil, err
	}
	if req.ThName != provider.Name() {
		return nil, xerr.ErrThirdPartyServiceNameNotMatch
	}
	rr, err := provider.QueryAudioReviewResult(ctx, req.Ext)
	if err != nil {
		return nil, err
	}
	res := &thirdpartypb.QueryAudioReviewResultRes{Status: rr.Status, Message: rr.RiskDescription}
	return res, nil
}

func (c intCtrl) QueryVideoReviewResult(ctx context.Context, req *thirdpartypb.QueryVideoReviewResultReq) (*thirdpartypb.QueryVideoReviewResultRes, error) {
	err := __checkReviewExtUniqReqId(req.Ext)
	if err != nil {
		return nil, err
	}
	if req.ThName != provider.Name() {
		return nil, xerr.ErrThirdPartyServiceNameNotMatch
	}
	rr, err := provider.QueryVideoReviewResult(ctx, req.Ext)
	if err != nil {
		return nil, err
	}
	res := &thirdpartypb.QueryVideoReviewResultRes{Status: rr.Status, Message: rr.RiskDescription}
	return res, nil
}
