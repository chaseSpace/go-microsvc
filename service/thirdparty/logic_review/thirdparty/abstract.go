package thirdparty

import (
	"context"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/thirdpartypb"
)

// ReviewAPI 在接入第三方审核服务时实现这个接口
type ReviewAPI interface {
	Name() string
	MustInit(config interface{})
	ReviewText(ctx context.Context, uid int64, text string, textType thirdpartypb.TextType, params *thirdpartypb.ReviewParamsExt) (*ReviewResult, error)
	ReviewImage(ctx context.Context, uid int64, uri string, imgType thirdpartypb.ImageType, params *thirdpartypb.ReviewParamsExt) (*ReviewResult, error)
	AsyncReviewAudio(ctx context.Context, uid int64, uri string, audioType thirdpartypb.AudioType, params *thirdpartypb.ReviewParamsExt) (*AsyncReviewResult, error)
	AsyncReviewVideo(ctx context.Context, uid int64, uri string, videoType thirdpartypb.VideoType, params *thirdpartypb.ReviewParamsExt) (*AsyncReviewResult, error)
	QueryAudioReviewResult(ctx context.Context, params *thirdpartypb.ReviewParamsExt) (*ReviewResult, error)
	QueryVideoReviewResult(ctx context.Context, params *thirdpartypb.ReviewParamsExt) (*ReviewResult, error)
}

// ReviewResult 同步审核请求成功结果
type ReviewResult struct {
	Status          commonpb.AIReviewStatus
	ReqId           string
	RiskLabel       string
	RiskDescription string
}

// AsyncReviewResult 异步审核请求成功结果
type AsyncReviewResult struct {
	ReqId             string // 本次请求ID，用于后续查询结果
	ThirdPartySvcName string // 使用哪个第三方服务，如 Shumei
}
