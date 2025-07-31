package logic_review

import (
	"context"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/commonpb"
	"microsvc/service/admin/dao"
	"unicode/utf8"

	"github.com/samber/lo"

	"github.com/pkg/errors"
)

type intCtrl struct {
}

var Int intCtrl

func (intCtrl) AddReview(ctx context.Context, req *adminpb.AddReviewReq) (res *adminpb.AddReviewRes, err error) {
	if utf8.RuneCountInString(req.Text) > 500 {
		return nil, errors.New("文本长度超出限制")
	}

	if !lo.Contains([]commonpb.ReviewStatus{
		commonpb.ReviewStatus_RS_Pending,
		commonpb.ReviewStatus_RS_AIPass,
		commonpb.ReviewStatus_RS_AIReject,
		commonpb.ReviewStatus_RS_Manual,
	}, req.Status) {
		return nil, errors.New("不支持录入的审核状态: " + req.Status.String())
	}

	if req.Type != commonpb.ReviewType_RT_Text && len(req.MediaUrls) == 0 {
		return nil, errors.New("请提供媒体资源URL")
	}

	switch req.Type {
	case commonpb.ReviewType_RT_Text:
		err = dao.ReviewDao.AddText(ctx, req)
	case commonpb.ReviewType_RT_Image:
		err = dao.ReviewDao.AddImage(ctx, req)
	case commonpb.ReviewType_RT_Video:
		err = dao.ReviewDao.AddVideo(ctx, req)
	case commonpb.ReviewType_RT_Audio:
		err = dao.ReviewDao.AddAudio(ctx, req)
	default:
		return nil, errors.New("无效的资源类型")
	}
	return &adminpb.AddReviewRes{}, err
}
