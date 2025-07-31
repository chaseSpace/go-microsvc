package thirdparty

import (
	"microsvc/enums"
	"microsvc/infra/svccli/rpc"
	"microsvc/infra/svccli/rpcext"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/thirdpartypb"
	deploy2 "microsvc/service/thirdparty/deploy"
	"microsvc/test/tbase"
	"microsvc/util/ufile"
	"path/filepath"
	"testing"

	"github.com/k0kubun/pp/v3"
)

func init() {
	tbase.TearUp(enums.SvcThirdparty, deploy2.ThirdpartyConf)
}

func TestSendSmsCode(t *testing.T) {
	defer tbase.TearDown()

	req := &thirdpartypb.SendSmsCodeIntReq{
		AreaCode: "86",
		Phone:    "18587904111",
		Scene:    commonpb.SmsCodeScene_SCS_SignUp,
		TestOnly: true,
	}
	rpc.Thirdparty().SendSmsCodeInt(tbase.TestCallCtxNoAuth, req)
}

func TestSendEmailCode(t *testing.T) {
	defer tbase.TearDown()

	req := &thirdpartypb.SendEmailCodeIntReq{
		Email:    "0@qq.com",
		Scene:    commonpb.EmailCodeScene_ECS_SignUp,
		TestOnly: true,
	}
	rpc.Thirdparty().SendEmailCodeInt(tbase.TestCallCtxNoAuth, req)
}

func TestVerifyEmailCode(t *testing.T) {
	defer tbase.TearDown()

	req := &thirdpartypb.VerifyEmailCodeReq{
		Base:       tbase.TestBaseExtReq,
		InputEmail: "0@qq.com",
		InputCode:  "589720",
		Scene:      commonpb.EmailCodeScene_ECS_ResetPasswd,
	}
	rpcext.Thirdparty().VerifyEmailCode(tbase.TestCallCtxNoAuth, req)
}

func TestOssUpload(t *testing.T) {
	defer tbase.TearDown()

	req := &thirdpartypb.OssUploadReq{
		Base: tbase.TestBaseExtReq,
		Buf:  []byte{0xFF, 0xD8, 0xFF},
		Type: commonpb.OSSUploadType_OUT_Avatar,
	}

	rpcext.Thirdparty().OssUpload(tbase.TestCallCtx, req)
}

func Test_LocalUpload(t *testing.T) {
	defer tbase.TearDown()
	req := &thirdpartypb.LocalUploadReq{
		Base:          tbase.TestBaseExtReq,
		FileBufBase64: ufile.MustReadToBase64(filepath.Clean("test\\testdata\\baidu.png")),
		BizType:       1,
	}

	r, e := rpcext.Thirdparty().LocalUpload(tbase.TestCallCtx, req)
	if e != nil {
		t.Fatal(e.Error())
	}
	pp.Println(r)
}
