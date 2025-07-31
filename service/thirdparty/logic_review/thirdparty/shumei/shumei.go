package shumei

import (
	"context"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/service/thirdparty/logic_review/thirdparty"
	"time"

	"github.com/parnurzeal/gorequest"
)

const (
	textUrl  = "https://api-text-bj.fengkongcloud.com/text/v4"
	imageUrl = "https://api-img-bj.fengkongcloud.com/image/v4"
	videoUrl = "https://api-video-bj.fengkongcloud.com/video/v4"
	audioUrl = "https://api-audio-bj.fengkongcloud.com/audio/v4"

	videoResultQueryUrl = "https://api-video-bj.fengkongcloud.com/video/query/v4"
)

var _ thirdparty.ReviewAPI = (*Shumei)(nil)

type Shumei struct {
	config *Config
	gor    *gorequest.SuperAgent
}

func (s *Shumei) MustInit(config interface{}) {
	s.config = config.(*Config)
	s.gor = gorequest.New().Timeout(time.Second * 10)
	s.gor.DoNotClearSuperAgent = true
}

func (s *Shumei) Name() string {
	return "Shumei"
}

// ReviewText 同步-文本审核: https://help.ishumei.com/docs/tj/text/newest/developDoc
func (s *Shumei) ReviewText(ctx context.Context, uid int64, text string, textType thirdpartypb.TextType, params *thirdpartypb.ReviewParamsExt) (*thirdparty.ReviewResult, error) {
	body := s.buildTextPayload(uid, text, textType)
	rr := TextResponse{}
	_, _, errs := s.gor.Post(textUrl).SendStruct(body).EndStruct(&rr)
	if len(errs) > 0 {
		errs = append(errs, xerr.ErrHTTPRequest)
		return nil, xerr.JoinErrors(errs...)
	}
	return rr.ToResult()
}

// ReviewImage 同步-图片审核: https://help.ishumei.com/docs/tj/image/versionV4/syncSingle/developDoc
func (s *Shumei) ReviewImage(ctx context.Context, uid int64, uri string, imgType thirdpartypb.ImageType, params *thirdpartypb.ReviewParamsExt) (*thirdparty.ReviewResult, error) {
	body := s.buildImagePayload(uid, uri, imgType)
	rr := ImageResponse{}
	_, _, errs := s.gor.Post(imageUrl).SendStruct(body).EndStruct(&rr)
	if len(errs) > 0 {
		errs = append(errs, xerr.ErrHTTPRequest)
		return nil, xerr.JoinErrors(errs...)
	}
	return rr.ToResult()
}

// AsyncReviewAudio 异步-语音审核
func (s *Shumei) AsyncReviewAudio(ctx context.Context, uid int64, uri string, audioType thirdpartypb.AudioType, params *thirdpartypb.ReviewParamsExt) (*thirdparty.AsyncReviewResult, error) {
	//TODO implement me
	panic("implement me")
}

// AsyncReviewVideo 异步-视频审核：https://help.ishumei.com/docs/tj/video/versionV4/requestInterface/developDoc
func (s *Shumei) AsyncReviewVideo(ctx context.Context, uid int64, uri string, videoType thirdpartypb.VideoType, params *thirdpartypb.ReviewParamsExt) (*thirdparty.AsyncReviewResult, error) {
	body := s.buildVideoPayload(uid, uri, videoType, params.UniqReqId.Val)
	rr := VideoResponse{}
	_, _, errs := s.gor.Post(videoUrl).SendStruct(body).EndStruct(&rr)
	if len(errs) > 0 {
		errs = append(errs, xerr.ErrHTTPRequest)
		return nil, xerr.JoinErrors(errs...)
	}
	return rr.ToResult(s.Name())
}

// QueryAudioReviewResult 查询语音审核结果
func (s *Shumei) QueryAudioReviewResult(ctx context.Context, params *thirdpartypb.ReviewParamsExt) (*thirdparty.ReviewResult, error) {
	panic(1)
}

// QueryVideoReviewResult 查询视频审核结果：https://help.ishumei.com/docs/tj/video/versionV4/queryInterface/developDoc
func (s *Shumei) QueryVideoReviewResult(ctx context.Context, params *thirdpartypb.ReviewParamsExt) (*thirdparty.ReviewResult, error) {
	body := &VideoQueryPayload{
		AccessKey: s.config.AccessKey,
		BtId:      params.UniqReqId.Val,
	}
	rr := VideoResultResp{}
	_, _, errs := s.gor.Post(videoResultQueryUrl).SendStruct(body).EndStruct(&rr)
	if len(errs) > 0 {
		errs = append(errs, xerr.ErrHTTPRequest)
		return nil, xerr.JoinErrors(errs...)
	}
	return rr.ToResult()
}
