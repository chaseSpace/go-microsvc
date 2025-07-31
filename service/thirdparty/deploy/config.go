package deploy

import (
	"errors"
	"microsvc/deploy"
	tencentOss "microsvc/service/thirdparty/logic_oss/thirdparty/tencent"
	shumeiReview "microsvc/service/thirdparty/logic_review/thirdparty/shumei"
	aliyunSms "microsvc/service/thirdparty/logic_sms/thirdparty/aliyun"
	tencentSms "microsvc/service/thirdparty/logic_sms/thirdparty/tencent"
)

type SvcConfig struct {
	deploy.CommConfig `mapstructure:"root"`

	Review struct {
		Shumei *shumeiReview.Config `mapstructure:"shumei"`
	} `mapstructure:"review"`

	Oss struct {
		LocalUploadDir string             `mapstructure:"local_upload_dir"`
		Tencent        *tencentOss.Config `mapstructure:"tencent"`
	} `mapstructure:"oss"`

	Sms struct {
		Tencent *tencentSms.Config `mapstructure:"tencent"`
		Aliyun  *aliyunSms.Config  `mapstructure:"aliyun"`
	} `mapstructure:"sms"`

	Email struct {
		SMTPHost string `mapstructure:"smtp_host"`
		SMTPPort int    `mapstructure:"smtp_port"`
		Sender   string `mapstructure:"sender"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
	}
}

func (s *SvcConfig) SelfCheck() error {
	if s.Oss.LocalUploadDir == "" {
		return errors.New("oss.local_upload_dir is empty")
	}
	return nil
}

var _ deploy.SvcConfImpl = new(SvcConfig)

var ThirdpartyConf = new(SvcConfig)
