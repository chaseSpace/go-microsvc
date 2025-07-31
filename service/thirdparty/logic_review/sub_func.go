package logic_review

import (
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/thirdpartypb"
)

func __checkReviewExtUniqReqId(ext *thirdpartypb.ReviewParamsExt) error {
	if ext == nil {
		return xerr.ErrParams.New("Field `ext` is required")
	}
	if ext.UniqReqId == nil || ext.UniqReqId.Val == "" {
		return xerr.ErrParams.New("Field `Ext.UniqReqId` is required")
	}
	return nil
}
