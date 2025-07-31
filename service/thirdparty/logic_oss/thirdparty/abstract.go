package thirdparty

import "context"

/*
OSS 使用注意事项

- 配置CDN域名，加速访问且降低OSS流量费
	-	开启CDN后，一般会设置存储桶的权限为私有读写
		然后设置CDN回源鉴权，腾讯OSS参考 https://cloud.tencent.com/document/product/436/18669#.E7.A7.81.E6.9C.89.E8.AF.BB.E5.AD.98.E5.82.A8.E6.A1.B6
- 在平台进行跨域配置，以允许来自web端的跨域请求
- 开启智能分层存储，减少OSS存储费
- 防盗刷：开启防盗链、配置流量预警
*/

type OssAPI interface {
	Name() string
	MustInit(config interface{})

	// UploadUserResource 上传用户资源
	// - fileNameWithPath 存储桶内的相对路径，建议规则：uid/666/avatar/20231024_110020.png （内部实现不做任何修改）
	UploadUserResource(ctx context.Context, ossPath string, buf []byte) (path, url string, err error)
	// UploadPublicResource 上传文件资源（参数含义与 UploadUserResource 一致）
	UploadPublicResource(ctx context.Context, ossPath string, buf []byte) (path, url string, err error)
}
