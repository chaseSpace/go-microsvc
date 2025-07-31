package consumer

import (
	"go.uber.org/zap"
	"microsvc/pkg/xlog"
)

type API interface {
	Init()
	ConsumerName() string
}

var registry = []API{
	consumerWithMicroSvc,
	consumerWithUser,
}

func Init() {
	_map := map[string]*struct{}{}
	for _, api := range registry {
		if _map[api.ConsumerName()] != nil {
			panic("duplicate consumer name: " + api.ConsumerName())
		}
		_map[api.ConsumerName()] = &struct{}{}
		api.Init()
		xlog.Debug("consumer initialized", zap.String("name", api.ConsumerName()))
	}
}
