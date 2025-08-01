package simple_sd

import (
	"context"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"microsvc/infra/sd/abstract"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/util/urand"
	"microsvc/xvendor/simple_sd"
	"time"
)

// SimpleSd is Single process sd implementation
type SimpleSd struct {
	serverPort int
	lastHash   string
	registry   map[string]*simple_sd.RegisterReq // svc -> id
}

func New(port int) *SimpleSd {
	return &SimpleSd{serverPort: port, registry: make(map[string]*simple_sd.RegisterReq)}
}

var _ abstract.ServiceDiscovery = (*SimpleSd)(nil)
var logPrefix = "simple_sd."

const (
	Name          = "simple_sd"
	httpResOkCode = 200

	registerPath    = "/service/register"
	deregisterPath  = "/service/deregister"
	discoveryPath   = "/service/discovery"
	healthCheckPath = "/service/health_check"
)

func (s *SimpleSd) getRequestUrl(path string) string {
	return fmt.Sprintf("http://localhost:%d%s", s.serverPort, path)
}

func (s *SimpleSd) Name() string {
	return Name
}

type httpRes struct {
	Code int // 200 OK
	Msg  string
	Data interface{} `json:"Data,omit_empty"`
}

func (s *SimpleSd) Register(service string, host string, port int, metadata map[string]string) error {
	if s.registry[service] != nil {
		return fmt.Errorf("already registered")
	}
	req := &simple_sd.RegisterReq{ServiceInstance: simple_sd.ServiceInstance{
		Id:       urand.Strings(4),
		Name:     service,
		IsUDP:    false,
		Host:     host,
		Port:     port,
		Metadata: metadata,
	}}
	res := new(httpRes)
	_, _, errs := gorequest.New().Post(s.getRequestUrl(registerPath)).SendStruct(req).EndStruct(res)
	if len(errs) > 0 {
		return errs[0]
	}
	if res.Code != httpResOkCode {
		return xerr.ErrInternal.New("register failed, got resp: %+v", res)
	}
	s.registry[service] = req
	return nil
}

func (s *SimpleSd) Deregister(service string) error {
	params := s.registry[service]
	if params == nil {
		return xerr.ErrInternal.New("never called Register")
	}
	type deregisterReq struct {
		Service string
		Id      string
	}
	req := &deregisterReq{
		Service: service,
		Id:      params.Id,
	}
	res := new(httpRes)
	_, _, errs := gorequest.New().Post(s.getRequestUrl(deregisterPath)).SendStruct(req).EndStruct(res)
	if len(errs) > 0 {
		return errs[0]
	}
	if res.Code != httpResOkCode {
		return xerr.ErrInternal.New("deregister failed, got resp: %+v", res)
	}
	delete(s.registry, params.Id)
	return nil
}

func (s *SimpleSd) Discover(ctx context.Context, serviceName string, block bool) ([]abstract.ServiceInstance, error) {
	req := &simple_sd.DiscoveryReq{
		Service:   serviceName,
		LastHash:  s.lastHash,
		WaitMaxMs: time.Minute.Milliseconds() * 2,
	}
	if !block {
		req.LastHash = ""
	}
	data := new(simple_sd.DiscoveryRspBody)
	res := &httpRes{Data: data}

	_, _, errs := gorequest.New().Post(s.getRequestUrl(discoveryPath)).SendStruct(req).EndStruct(res)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	if res.Code != httpResOkCode {
		return nil, xerr.ErrInternal.New("discovery failed, got resp: %+v", res)
	}
	s.lastHash = data.Hash
	return lo.Map(data.Instances, func(item simple_sd.ServiceInstance, index int) abstract.ServiceInstance {
		return abstract.ServiceInstance{
			ID:       item.Id,
			Name:     item.Name,
			IsUDP:    item.IsUDP,
			Host:     item.Host,
			Port:     item.Port,
			Metadata: item.Metadata,
		}
	}), nil
}

func (s *SimpleSd) HealthCheck(ctx context.Context, service string) error {
	params := s.registry[service]
	if params == nil {
		return xerr.ErrInternal.New("never called Register")
	}
	req := &simple_sd.HealthCheckReq{
		Service: service,
		Id:      params.Id,
	}
	rspBody := new(simple_sd.HealthCheckRspBody)
	res := &httpRes{Data: rspBody}
	_, _, errs := gorequest.New().Post(s.getRequestUrl(healthCheckPath)).SendStruct(req).EndStruct(res)
	if len(errs) > 0 {
		return errs[0]
	}
	if res.Code != httpResOkCode {
		return xerr.ErrInternal.New("health check failed, got resp: %+v", res)
	}
	if !rspBody.Registered {
		xlog.Warn(fmt.Sprintf(logPrefix+"HealthCheck: service [%s - id:%s] offline, do re-register now", service, params.Id))
		delete(s.registry, params.Name)
		err := s.Register(params.Name, params.Host, params.Port, params.Metadata)
		return err
	}
	return nil
}

func (s *SimpleSd) Stop() {
	for _, r := range s.registry {
		err := s.Deregister(r.Name)
		if err != nil {
			xlog.Error(logPrefix+"Stop: deregister fail", zap.Error(err), zap.String("svc", r.Name))
		} else {
			xlog.Info(logPrefix+"Stop: deregister success", zap.String("svc", r.Name))
		}
	}
}
