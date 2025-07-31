package deploy

import (
	"microsvc/deploy"
)

type SvcConfig struct {
	deploy.CommConfig `mapstructure:"root"`

	Kimi struct {
		APIKey string `mapstructure:"api_key"`
	} `mapstructure:"kimi"`
	Tongyi struct {
		APIKey string `mapstructure:"api_key"`
	} `mapstructure:"tongyi"`
}

var _ deploy.SvcConfImpl = new(SvcConfig)

var CrontaskConf = new(SvcConfig)
