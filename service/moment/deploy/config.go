package deploy

import (
	"microsvc/deploy"
)

type SvcConfig struct {
	deploy.CommConfig `mapstructure:"root"`
}

var _ deploy.SvcConfImpl = new(SvcConfig)

var MomentConf = new(SvcConfig)
