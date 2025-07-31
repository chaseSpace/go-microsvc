package handler

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/service/thirdparty/logic_email"
	"microsvc/service/thirdparty/logic_oss"
	"microsvc/service/thirdparty/logic_review"
)

var Ctrl thirdpartypb.ThirdpartyExtServer = new(ctrl)

type ctrl struct{}

func (c ctrl) VerifyEmailCode(ctx context.Context, req *thirdpartypb.VerifyEmailCodeReq) (*thirdpartypb.VerifyEmailCodeRes, error) {
	return logic_email.Ext.VerifyEmailCode(ctx, req)
}

func (c ctrl) OssUpload(ctx context.Context, req *thirdpartypb.OssUploadReq) (*thirdpartypb.OssUploadRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic_oss.Ext.OssUpload(ctx, caller, req)
}

func (c ctrl) SyncReviewImageExt(ctx context.Context, req *thirdpartypb.SyncReviewImageExtReq) (*thirdpartypb.SyncReviewImageExtRes, error) {
	caller := auth.ExtractSvcUser(ctx)
	return logic_review.Ext.SyncReviewImageExt(ctx, caller.Uid, req)
}

func (c ctrl) LocalUpload(ctx context.Context, req *thirdpartypb.LocalUploadReq) (*thirdpartypb.LocalUploadRes, error) {
	return logic_oss.Ext.LocalUpload(ctx, req)
}
