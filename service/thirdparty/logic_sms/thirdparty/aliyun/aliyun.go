package aliyun

import (
	"context"
	"fmt"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/service/thirdparty/logic_sms/thirdparty"
	"microsvc/util"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sms "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var _ thirdparty.SmsAPI = (*SmsImpl)(nil)

type SmsImpl struct {
	client *sms.Client
	config *Config
	gr     *zap.Logger
}

func (s *SmsImpl) Name() string {
	return "AliyunSms"
}

func (s *SmsImpl) MustInit(config interface{}) {
	s.config = config.(*Config)
	cc := &openapi.Config{
		AccessKeyId:     tea.String(s.config.AccessKeyID),
		AccessKeySecret: tea.String(s.config.AccessKeySecret),
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
	}

	var err error
	s.client, err = sms.NewClient(cc)
	util.AssertNilErr(err)

	s.gr = xlog.WithFields(zap.String("SDK", s.Name()))
}

func (s *SmsImpl) newRequest(areaCode string, phone, code string) (*sms.SendSmsRequest, string) {
	phoneVar := fmt.Sprintf("+%s%s", areaCode, phone)

	signName := s.config.SignName
	templateCode := s.config.TemplateCode

	// 海外短信
	if areaCode != "86" {
		signName = s.config.OverseasSignName // 阿里云的国外短信也使用签名
		templateCode = s.config.OverseasTemplateCode
	}

	req := &sms.SendSmsRequest{
		PhoneNumbers:  tea.String(phoneVar),
		SignName:      tea.String(signName),
		TemplateCode:  tea.String(templateCode),
		TemplateParam: tea.String(util.ToJsonStr(map[string]string{"code": code})),
		//SmsUpExtendCode: nil,
	}

	return req, phoneVar
}

func (s *SmsImpl) sendSmsCode(ctx context.Context, areaCode, phone, code, hint string) error {
	request, phoneVar := s.newRequest(areaCode, phone, code)

	// *Timeout 都是毫秒单位
	// API Doc：https://next.api.aliyun.com/document/Dysmsapi/2017-05-25/SendSms
	response, err := s.client.SendSmsWithOptions(request, &service.RuntimeOptions{ReadTimeout: tea.Int(3000)})
	if err != nil {
		s.gr.Error(hint+"-Failed", zap.Error(err), zap.Any("request", request))
		var terr *tea.SDKError
		if errors.As(err, &terr) {
			return xerr.ErrThirdParty.AppendMsg(s.Name() + ": " + tea.StringValue(terr.Message))
		}
		return xerr.ErrThirdParty.AppendMsg(s.Name() + " err")
	}
	s.gr.Info(hint+"-OK", zap.Any("response", response), zap.String("phoneVar", phoneVar), zap.String("code", code))
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
