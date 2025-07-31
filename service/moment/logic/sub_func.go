package logic

import (
	"context"
	"fmt"
	"microsvc/bizcomm/commadmin"
	"microsvc/bizcomm/commthirdparty"
	"microsvc/enums"
	"microsvc/infra/svccli/rpc"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/momentpb"
	"microsvc/protocol/svc/thirdpartypb"
	"unicode/utf8"

	"github.com/pkg/errors"
)

func __checkNewMoment(ctx context.Context, req *momentpb.CreateMomentReq) error {
	_lenText := utf8.RuneCountInString(req.Text)
	if _lenText > enums.MomentMaxTextLength {
		return xerr.ErrMomentTextTooLong
	}
	switch req.Type {
	case momentpb.MomentType_MT_Text:
		if _lenText == 0 {
			return xerr.ErrParams.New("不能发空文本动态")
		}
	case momentpb.MomentType_MT_Audio:
		if len(req.MediaUrls) != 1 {
			return xerr.ErrParams.New("需要上传仅1段音频哦~")
		}
	case momentpb.MomentType_MT_Image:
		if len(req.MediaUrls) == 0 {
			return xerr.ErrParams.New("请上传至少1张图片哦~")
		}
		if len(req.MediaUrls) > 9 {
			return xerr.ErrParams.New("不能超过9张图片哦~")
		}
	case momentpb.MomentType_MT_Video:
		if len(req.MediaUrls) != 1 {
			return xerr.ErrParams.New("需要上传仅1个视频哦~")
		}
	default:
		return xerr.ErrMomentTypeNotFound
	}
	return nil
}

func __execAIReview(ctx context.Context, uid int64, req *momentpb.CreateMomentReq) (result commthirdparty.ReviewResult, taskID string, err error) {
	sw, err := commadmin.SwitchCenterGetOne(ctx, commadmin.SwitchKeyAIReviewUserNewMoment)
	if err != nil {
		return nil, "", err
	}

	if sw.IsClose() {
		return
	}

	// 只有文字和图片是同步审核
	switch req.Type {
	case momentpb.MomentType_MT_Text:
		result, err = __AIReviewText(ctx, uid, req)
	case momentpb.MomentType_MT_Image:
		result, err = __AIReviewImg(ctx, uid, req)
	case momentpb.MomentType_MT_Audio:
		taskID, err = __AIReviewAudio(ctx, uid, req)
	case momentpb.MomentType_MT_Video:
		taskID, err = __AIReviewVideo(ctx, uid, req)
	}
	return
}

func __AIReviewText(ctx context.Context, uid int64, req *momentpb.CreateMomentReq) (result commthirdparty.ReviewResult, err error) {
	res, err := rpc.Thirdparty().SyncReviewText(ctx, &thirdpartypb.SyncReviewTextReq{
		Uid:  uid,
		Text: req.Text,
		Type: thirdpartypb.TextType_TT_Moment,
	})
	return res, errors.Wrap(err, "审核文本失败")
}

func __AIReviewImg(ctx context.Context, uid int64, req *momentpb.CreateMomentReq) (result commthirdparty.ReviewResult, err error) {
	//  NOTE：由于是多次调用机审接口，这里可能存在潜在的接口性能问题
	// TODO: 优化方向：客户端上传时调用后端审核接口，而不是在发布时审核
	for i, uri := range req.MediaUrls {
		res, err := rpc.Thirdparty().SyncReviewImage(ctx, &thirdpartypb.SyncReviewImageReq{
			Uid:  uid,
			Uri:  uri,
			Type: thirdpartypb.ImageType_IT_Moment,
		})
		if err != nil {
			return res, errors.Wrap(err, fmt.Sprintf("第%d张图片审核失败", i+1))
		}
		if res.GetStatus() == commonpb.AIReviewStatus_ARS_Reject {
			return res, errors.Wrap(err, fmt.Sprintf("第%d张图片违规", i+1))
		}
	}
	return &thirdpartypb.SyncReviewTextRes{Status: commonpb.AIReviewStatus_ARS_Pass}, nil
}

func __AIReviewAudio(ctx context.Context, uid int64, req *momentpb.CreateMomentReq) (taskID string, err error) {
	res, err := rpc.Thirdparty().AsyncReviewAudio(ctx, &thirdpartypb.AsyncReviewAudioReq{
		Uid:  uid,
		Uri:  req.MediaUrls[0],
		Type: thirdpartypb.AudioType_AT_Moment,
	})
	if err != nil {
		return "", errors.Wrap(err, "异步审核音频失败")
	}
	return res.ReqId, nil
}

func __AIReviewVideo(ctx context.Context, uid int64, req *momentpb.CreateMomentReq) (taskID string, err error) {
	res, err := rpc.Thirdparty().AsyncReviewVideo(ctx, &thirdpartypb.AsyncReviewVideoReq{
		Uid:  uid,
		Uri:  req.MediaUrls[0],
		Type: thirdpartypb.VideoType_VT_Moment,
	})
	if err != nil {
		return "", errors.Wrap(err, "异步审核视频失败")
	}
	return res.ReqId, nil
}

func __AddAdminReview(ctx context.Context, uid int64, req *momentpb.CreateMomentReq, mid int64, taskID string, status commonpb.AIReviewStatus) error {
	req2 := &adminpb.AddReviewReq{
		Uid:       uid,
		Type:      0, // 下面填充
		Text:      req.Text,
		MediaUrls: req.MediaUrls,
		Status:    0, // 下面填充
		BizType:   commonpb.BizType_RBT_Moment,
		BizUniqId: mid,
		ThTaskId:  taskID,
	}
	// AI审核状态 转换为 通用审核状态
	switch status {
	case commonpb.AIReviewStatus_ARS_Pass:
		req2.Status = commonpb.ReviewStatus_RS_AIPass
	case commonpb.AIReviewStatus_ARS_Reject:
		req2.Status = commonpb.ReviewStatus_RS_AIReject
	case commonpb.AIReviewStatus_ARS_Review:
		req2.Status = commonpb.ReviewStatus_RS_Manual
	default:
		return xerr.ErrParams.New("无效的AI审核状态")
	}

	// 动态内的媒体类型 转换为 通用媒体类型
	switch req.Type {
	case momentpb.MomentType_MT_Text:
		req2.Type = commonpb.ReviewType_RT_Text
	case momentpb.MomentType_MT_Image:
		req2.Type = commonpb.ReviewType_RT_Image
	case momentpb.MomentType_MT_Video:
		req2.Type = commonpb.ReviewType_RT_Video
	case momentpb.MomentType_MT_Audio:
		req2.Type = commonpb.ReviewType_RT_Audio
	default:
		return xerr.ErrParams.New("无效的动态类型")
	}
	_, err := rpc.Admin().AddReview(ctx, req2)
	return err
}

func __convertAIReviewStatus(s commonpb.AIReviewStatus) (momentpb.ReviewStatus, error) {
	s2 := map[commonpb.AIReviewStatus]momentpb.ReviewStatus{
		commonpb.AIReviewStatus_ARS_Pass:   momentpb.ReviewStatus_RS_Pass,
		commonpb.AIReviewStatus_ARS_Reject: momentpb.ReviewStatus_RS_Reject,
	}[s]
	if s2 == 0 {
		return 0, xerr.ErrParams.New("不支持的审核状态：" + s.String())
	}
	return s2, nil
}
