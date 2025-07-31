package thirdparty

import (
	"microsvc/enums"
	"microsvc/infra/svccli/rpc"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/service/thirdparty/deploy"
	"microsvc/test/tbase"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifySmsCodeInt(t *testing.T) {
	tbase.TearUp(enums.SvcThirdparty, deploy.ThirdpartyConf)
	defer tbase.TearDown()

	// case-1 未提前获取code
	r, err := rpc.Thirdparty().VerifySmsCodeInt(tbase.TestCallCtx, &thirdpartypb.VerifySmsCodeIntReq{
		Uid:      0,
		AreaCode: "86",
		Phone:    "111",
		Code:     "111",
		Scene:    commonpb.SmsCodeScene_SCS_SignUp,
	})
	assert.Nil(t, err)
	assert.False(t, r.Pass)
}

func TestVerifyEmailCodeInt(t *testing.T) {
	tbase.TearUp(enums.SvcThirdparty, deploy.ThirdpartyConf)
	defer tbase.TearDown()

	rpc.Thirdparty().VerifyEmailCodeInt(tbase.TestCallCtx, &thirdpartypb.VerifyEmailCodeIntReq{
		Email: "random2035@qq.com",
		Code:  "297536",
		Scene: commonpb.EmailCodeScene_ECS_SignUp,
	})
}

func TestSyncReviewText(t *testing.T) {
	tbase.TearUp(enums.SvcThirdparty, deploy.ThirdpartyConf)
	defer tbase.TearDown()

	rpc.Thirdparty().SyncReviewText(tbase.TestCallCtx, &thirdpartypb.SyncReviewTextReq{
		Uid:  1,
		Text: "xxx",
		Type: thirdpartypb.TextType_TT_Nickname,
		Ext:  nil,
	})
}

func TestSyncReviewImage(t *testing.T) {
	tbase.TearUp(enums.SvcThirdparty, deploy.ThirdpartyConf)
	defer tbase.TearDown()

	rpc.Thirdparty().SyncReviewImage(tbase.TestCallCtx, &thirdpartypb.SyncReviewImageReq{
		Uid:  1,
		Uri:  "http://xxx.com/x.png",
		Type: thirdpartypb.ImageType_IT_Avatar,
		Ext:  nil,
	})
}

func TestAsyncReviewVideo(t *testing.T) {
	tbase.TearUp(enums.SvcThirdparty, deploy.ThirdpartyConf)
	defer tbase.TearDown()

	rpc.Thirdparty().AsyncReviewVideo(tbase.TestCallCtx, &thirdpartypb.AsyncReviewVideoReq{
		Uid:  1,
		Uri:  "http://xxx.com/x.png",
		Type: thirdpartypb.VideoType_VT_IM,
		Ext: &thirdpartypb.ReviewParamsExt{
			UniqReqId: &thirdpartypb.ReviewParamsExt_UniqReqId{Val: "1"},
			Sex:       nil,
		},
	})
}

func TestQueryVideoReviewResult(t *testing.T) {
	tbase.TearUp(enums.SvcThirdparty, deploy.ThirdpartyConf)
	defer tbase.TearDown()

	rpc.Thirdparty().QueryVideoReviewResult(tbase.TestCallCtx, &thirdpartypb.QueryVideoReviewResultReq{
		ThName: "Shumei",
		Ext: &thirdpartypb.ReviewParamsExt{
			UniqReqId: &thirdpartypb.ReviewParamsExt_UniqReqId{Val: "1"},
			Sex:       &thirdpartypb.ReviewParamsExt_Sex{Val: commonpb.Sex_Male},
		},
	})
}
