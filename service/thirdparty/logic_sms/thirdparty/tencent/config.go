package tencent

/*
**腾讯云短信服务**

阅读文档：https://cloud.tencent.com/document/product/382/43199
来看懂配置
*/

type Config struct {
	SecretId  string `mapstructure:"secret_id"`
	SecretKey string `mapstructure:"secret_key"`

	AppID string `mapstructure:"app_id"` // 短信应用id，不分国家

	// 国内短信的签名&模板
	SignName   string `mapstructure:"sign_name"`   // 签名
	TemplateID string `mapstructure:"template_id"` // 模板id

	// 海外短信的签名&模板
	OverseasSenderID   string `mapstructure:"overseas_sender_id"`   // 海外发送方id，可选，需要申请（每个国家）
	OverseasSignName   string `mapstructure:"overseas_sign_name"`   // 海外签名
	OverseasTemplateID string `mapstructure:"overseas_template_id"` // 海外模板
}
