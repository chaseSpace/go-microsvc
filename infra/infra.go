package infra

import (
	"microsvc/deploy"
	"microsvc/pkg/xlog"

	"go.uber.org/zap"
)

type initFunc func(cc *deploy.XConfig, onEnd func(must bool, err error))

func Setup(initFn ...initFunc) {
	for _, fn := range initFn {
		fn(deploy.XConf, func(must bool, err error) {
			if must && err != nil {
				panic(err)
			}
			if err != nil {
				xlog.Error("infra.Setup err", zap.Error(err))
			}
		})
	}
}
