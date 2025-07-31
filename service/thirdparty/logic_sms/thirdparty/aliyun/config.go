package aliyun

/*
**阿里云短信服务（使用最新的V2 SDK）**
https://next.api.aliyun.com/api-tools/sdk/Dysmsapi?version=2017-05-25&language=go-tea&tab=primer-doc

- 注意在【系统设置】模块下开启防盗刷、监控预警策略
*/

type Config struct {
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`

	// 国内短信的签名&模板
	SignName     string `mapstructure:"sign_name"`
	TemplateCode string `mapstructure:"template_id"`

	// 阿里云将海外的senderID也归为签名管理
	OverseasSignName     string `mapstructure:"overseas_sign_name"`
	OverseasTemplateCode string `mapstructure:"overseas_template_code"`
}
