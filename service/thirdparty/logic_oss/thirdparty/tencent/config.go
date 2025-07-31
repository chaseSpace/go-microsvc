package tencent

/*
**腾讯云OSS服务**

阅读文档：https://cloud.tencent.com/document/product/436/6222
SDK文档：https://cloud.tencent.com/document/product/436/31215
来看懂配置
*/

type Config struct {
	SecretId  string `mapstructure:"secret_id"`
	SecretKey string `mapstructure:"secret_key"`

	// 存储桶地址示例：https://examplebucket-1250000000.cos.ap-guangzhou.myqcloud.com
	// 查看存储桶地址：https://console.cloud.tencent.com/cos/bucket
	UserBucketURL   string `mapstructure:"user_bucket_url"`   // 用户存储桶地址，存储用户常见资源，如头像/背景等
	PublicBucketURL string `mapstructure:"public_bucket_url"` // 公共存储同地址，存储公开资源，如宣传页/banner等
}
