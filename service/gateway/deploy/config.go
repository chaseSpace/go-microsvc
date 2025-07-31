package deploy

import (
	"microsvc/deploy"
	"microsvc/util/urand"
)

// SvcConfig 每个服务特有的配置结构
type SvcConfig struct {
	deploy.CommConfig `mapstructure:"root"`
	HttpPort          int `mapstructure:"http_port"`

	Generated Generated `mapstructure:"-"`    // 自生成的配置
	Cors      Cors      `mapstructure:"cors"` // 跨域
}

// GwID shortcut of GatewayUniqID
func (c *SvcConfig) GwID() string {
	return c.Generated.GatewayUniqID
}

var _ deploy.SvcConfImpl = new(SvcConfig)

var GatewayConf = &SvcConfig{
	Generated: struct{ GatewayUniqID string }{GatewayUniqID: urand.Strings(4, true)},
}

// ------------------------------------

type Generated struct {
	GatewayUniqID string // 网关ID
}

type Cors struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
}
