package deploy

import (
	"microsvc/enums"
)

type SvcListenPortSetter interface {
	GetSvc() string
	SetGRPC(int)
	SetHTTP(int)
}

type RegisterSvc interface {
	RegGRPCBase() (name string, addr string, port int)
	RegGRPCMeta() map[string]string
}

type SvcConfImpl interface {
	GetLogLevel() string
	OverrideLogLevel(string)
	Name() enums.Svc
	SelfCheck() error
}

type CommConfig struct {
	Svc          enums.Svc `mapstructure:"svc"`
	LogLevel     string    `mapstructure:"log_level"`
	DisableCache bool      `mapstructure:"disable_cache"`
}

func (s *CommConfig) Name() enums.Svc {
	return s.Svc
}

func (s *CommConfig) GetLogLevel() string {
	return s.LogLevel
}

func (s *CommConfig) OverrideLogLevel(lv string) {
	s.LogLevel = lv
}

func (s *CommConfig) SelfCheck() error {
	return nil
}
