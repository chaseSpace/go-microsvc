package tencent

import (
	"bytes"
	"context"
	"microsvc/pkg/xlog"
	"microsvc/service/thirdparty/logic_oss/thirdparty"
	"net/http"
	"net/url"
	"strings"

	"github.com/tencentyun/cos-go-sdk-v5"
	"go.uber.org/zap"
)

/*
注意：在这一层不要包含用户相关的逻辑
*/

var _ thirdparty.OssAPI = (*OssImpl)(nil)

type OssImpl struct {
	userBucketClient   *cos.Client
	publicBucketClient *cos.Client
	config             *Config
	gr                 *zap.Logger
}

func (s *OssImpl) Name() string {
	return "TencentOss"
}

func (s *OssImpl) MustInit(config interface{}) {
	s.config = config.(*Config)

	transport := &cos.AuthorizationTransport{
		SecretID:  s.config.SecretId,
		SecretKey: s.config.SecretKey,
	}

	userBucketURL, _ := url.Parse(s.config.UserBucketURL)
	publicBucketURL, _ := url.Parse(s.config.PublicBucketURL)
	su, _ := url.Parse("https://service.cos.myqcloud.com")

	b := &cos.BaseURL{BucketURL: userBucketURL, ServiceURL: su}
	b2 := &cos.BaseURL{BucketURL: publicBucketURL, ServiceURL: su}

	s.userBucketClient = cos.NewClient(b, &http.Client{Transport: transport})
	s.publicBucketClient = cos.NewClient(b2, &http.Client{Transport: transport})

	s.gr = xlog.WithFields(zap.String("SDK", s.Name()))
}

func (s *OssImpl) UploadUserResource(ctx context.Context, ossPath string, buf []byte) (path, url string, err error) {
	// Put接口最大支持5GB，response无内容
	// https://cloud.tencent.com/document/product/436/7749#.E5.93.8D.E5.BA.94
	_, err = s.userBucketClient.Object.Put(ctx, ossPath, bytes.NewReader(buf), nil)
	return ossPath, strings.Join([]string{s.config.UserBucketURL, ossPath}, "/"), err
}

func (s *OssImpl) UploadPublicResource(ctx context.Context, ossPath string, buf []byte) (path, url string, err error) {
	_, err = s.publicBucketClient.Object.Put(ctx, ossPath, bytes.NewReader(buf), nil)
	return ossPath, strings.Join([]string{s.config.UserBucketURL, ossPath}, "/"), err
}
