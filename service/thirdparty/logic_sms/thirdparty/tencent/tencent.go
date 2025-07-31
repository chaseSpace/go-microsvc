package tencent

import (
	"context"
	"fmt"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/service/thirdparty/logic_sms/thirdparty"

	"github.com/pkg/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	terrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"go.uber.org/zap"
)

var _ thirdparty.SmsAPI = (*SmsImpl)(nil)

type SmsImpl struct {
	client *sms.Client
	config *Config
	gr     *zap.Logger
}

func (s *SmsImpl) Name() string {
	//TODO implement me
	return "TencentSms"
}

func (s *SmsImpl) MustInit(config interface{}) {
	s.config = config.(*Config)

	credential := common.NewCredential(s.config.SecretId, s.config.SecretKey)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 10
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"

	s.client, _ = sms.NewClient(credential, "ap-guangzhou", cpf)

	s.gr = xlog.WithFields(zap.String("SDK", s.Name()))
}

func (s *SmsImpl) newRequest(ctx context.Context, areaCode, phone, code string) (*sms.SendSmsRequest, string) {
	request := sms.NewSendSmsRequest()
	request.BaseRequest.SetContext(ctx)
	request.SmsSdkAppId = common.StringPtr(s.config.AppID)

	phoneVar := fmt.Sprintf("+%s%s", areaCode, phone)
	request.TemplateParamSet = common.StringPtrs([]string{code})
	request.PhoneNumberSet = common.StringPtrs([]string{phoneVar})

	signName := s.config.SignName
	templateID := s.config.TemplateID

	// 海外短信
	if areaCode != "86" {
		signName = s.config.OverseasSignName
		templateID = s.config.OverseasTemplateID
		request.SenderId = common.StringPtr(s.config.OverseasSenderID)
	}

	request.SignName = common.StringPtr(signName)
	request.TemplateId = common.StringPtr(templateID)
	return request, phoneVar
}

func (s *SmsImpl) sendSmsCode(ctx context.Context, areaCode, phone, code, hint string) error {
	request, phoneVar := s.newRequest(ctx, areaCode, phone, code)
	response, err := s.client.SendSms(request)
	if err != nil {
		s.gr.Error(hint+"-Failed", zap.Error(err), zap.Any("request", request))
		var terr *terrors.TencentCloudSDKError
		if errors.As(err, &terr) {
			return xerr.ErrThirdParty.AppendMsg(s.Name() + ": " + terr.Message)
		}
		return xerr.ErrThirdParty.AppendMsg(s.Name() + " err")
	}
	s.gr.Info(hint+"-OK", zap.Any("response", response.Response), zap.String("phoneVar", phoneVar), zap.String("code", code))
	return nil
}

// SendDomesticSmsCode 发送国内短信
func (s *SmsImpl) SendDomesticSmsCode(ctx context.Context, phone, code string) error {
	return s.sendSmsCode(ctx, "86", phone, code, "SendDomesticSmsCode")
}

// SendOverseasSmsCode 发送海外短信
func (s *SmsImpl) SendOverseasSmsCode(ctx context.Context, areaCode, phone, code string) error {
	return s.sendSmsCode(ctx, areaCode, phone, code, "SendOverseasSmsCode")
}
