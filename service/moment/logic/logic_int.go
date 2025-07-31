package logic

import (
	"context"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/momentpb"
	"microsvc/service/moment/dao"
	"time"
)

type intCtrl struct{}

var Int = intCtrl{}

func (i intCtrl) UpdateReviewStatus(ctx context.Context, req *momentpb.UpdateReviewStatusReq) (*momentpb.UpdateReviewStatusRes, error) {
	if req.Uid < 1 || req.Mid < 1 {
		return nil, xerr.ErrParams.New("无效的UID(%d)或MID(%d)", req.Uid, req.Mid)
	}
	passAt := int64(0)
	switch req.Status {
	case momentpb.ReviewStatus_RS_Pass:
		passAt = time.Now().UnixMilli()
	case momentpb.ReviewStatus_RS_Reject:
	default:
		return nil, xerr.ErrParams.New("不支持的status: " + req.Status.String())
	}
	err := dao.MomentDao.UpdateReviewStatus(ctx, req.Uid, req.Mid, req.Status, passAt)
	if err != nil {
		return nil, err
	}
	// todo 发送系统通知给用户
	return &momentpb.UpdateReviewStatusRes{}, err
}
