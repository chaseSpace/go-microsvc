package logic_oss

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/protocol/svc/thirdpartypb"
)

type ctrl struct {
}

var Ext ctrl

func (c ctrl) OssUpload(ctx context.Context, caller *auth.SvcCaller, req *thirdpartypb.OssUploadReq) (*thirdpartypb.OssUploadRes, error) {
	err := checkUploadParams(ctx, caller.Uid, req.Type, req.Buf)
	if err != nil {
		return nil, err
	}
	path, url, err := uploadResource(ctx, caller.Uid, req.Type, req.Buf)
	return &thirdpartypb.OssUploadRes{Path: path, Url: url}, err
}

func (c ctrl) LocalUpload(ctx context.Context, req *thirdpartypb.LocalUploadReq) (*thirdpartypb.LocalUploadRes, error) {
	uid := auth.GetAuthUID(ctx)
	path, accessUri, err := localUpload(ctx, uid, req.BizType, req.FileBufBase64)
	if err != nil {
		return nil, err
	}
	return &thirdpartypb.LocalUploadRes{Path: path, AccessUri: accessUri}, nil
}
