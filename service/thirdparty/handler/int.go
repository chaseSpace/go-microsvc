package handler

import (
	"context"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/service/thirdparty/logic_email"
	"microsvc/service/thirdparty/logic_oss"
	"microsvc/service/thirdparty/logic_review"
	"microsvc/service/thirdparty/logic_sms"
)

var IntCtrl thirdpartypb.ThirdpartyIntServer = new(intCtrl)

type intCtrl struct {
}

func (c intCtrl) SendSmsCodeInt(ctx context.Context, req *thirdpartypb.SendSmsCodeIntReq) (*thirdpartypb.SendSmsCodeIntRes, error) {
	return logic_sms.Int.SendSmsCodeInt(ctx, req)
}

func (c intCtrl) SendEmailCodeInt(ctx context.Context, req *thirdpartypb.SendEmailCodeIntReq) (*thirdpartypb.SendEmailCodeIntRes, error) {
	return logic_email.Int.SendEmailCodeInt(ctx, req)
}

func (intCtrl) VerifySmsCodeInt(ctx context.Context, req *thirdpartypb.VerifySmsCodeIntReq) (*thirdpartypb.VerifySmsCodeIntRes, error) {
	return logic_sms.Int.VerifySmsCodeInt(ctx, req)
}

func (c intCtrl) VerifyEmailCodeInt(ctx context.Context, req *thirdpartypb.VerifyEmailCodeIntReq) (*thirdpartypb.VerifyEmailCodeIntRes, error) {
	return logic_email.Int.VerifyEmailCodeInt(ctx, req)
}

func (intCtrl) SyncReviewText(ctx context.Context, req *thirdpartypb.SyncReviewTextReq) (*thirdpartypb.SyncReviewTextRes, error) {
	if req.Ext == nil {
		req.Ext = &thirdpartypb.ReviewParamsExt{}
	}
	return logic_review.Int.SyncReviewText(ctx, req)
}

func (intCtrl) SyncReviewImage(ctx context.Context, req *thirdpartypb.SyncReviewImageReq) (*thirdpartypb.SyncReviewImageRes, error) {
	if req.Ext == nil {
		req.Ext = &thirdpartypb.ReviewParamsExt{}
	}
	return logic_review.Int.SyncReviewImage(ctx, req)
}

func (c intCtrl) AsyncReviewAudio(ctx context.Context, req *thirdpartypb.AsyncReviewAudioReq) (*thirdpartypb.AsyncReviewAudioRes, error) {
	//TODO implement me
	panic("implement me")
}

func (c intCtrl) AsyncReviewVideo(ctx context.Context, req *thirdpartypb.AsyncReviewVideoReq) (*thirdpartypb.AsyncReviewVideoRes, error) {
	return logic_review.Int.AsyncReviewVideo(ctx, req)
}

func (c intCtrl) QueryAudioReviewResult(ctx context.Context, req *thirdpartypb.QueryAudioReviewResultReq) (*thirdpartypb.QueryAudioReviewResultRes, error) {
	return logic_review.Int.QueryAudioReviewResult(ctx, req)
}

func (c intCtrl) QueryVideoReviewResult(ctx context.Context, req *thirdpartypb.QueryVideoReviewResultReq) (*thirdpartypb.QueryVideoReviewResultRes, error) {
	return logic_review.Int.QueryVideoReviewResult(ctx, req)
}

func (c intCtrl) LocalUploadInt(ctx context.Context, req *thirdpartypb.LocalUploadIntReq) (*thirdpartypb.LocalUploadIntRes, error) {
	return logic_oss.Int.LocalUpload(ctx, req)
}
