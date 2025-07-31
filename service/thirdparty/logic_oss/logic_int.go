package logic_oss

import (
	"context"
	"microsvc/bizcomm/auth"
	"microsvc/protocol/svc/thirdpartypb"
)

type intCtrl struct{}

var Int = intCtrl{}

func (c intCtrl) LocalUpload(ctx context.Context, req *thirdpartypb.LocalUploadIntReq) (*thirdpartypb.LocalUploadIntRes, error) {
	uid := auth.GetAuthUID(ctx)
	path, accessUri, err := localUpload(ctx, uid, req.BizType, req.FileBufBase64)
	if err != nil {
		return nil, err
	}
	return &thirdpartypb.LocalUploadIntRes{Path: path, AccessUri: accessUri}, nil
}
