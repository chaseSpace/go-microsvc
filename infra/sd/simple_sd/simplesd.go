package simple_sd

import (
	"context"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"microsvc/infra/sd/spec"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/util/urand"
	"microsvc/xvendor/simple_sd"
	"sync"
	"time"
)

// SimpleSd is Single process sd implementation
type SimpleSd struct {
	serverPort int
	lastHash   string
	registry   map[string]*simple_sd.RegisterReq // svc -> id
	sync.RWMutex
}

func New(port int) *SimpleSd {
	return &SimpleSd{serverPort: port, registry: make(map[string]*simple_sd.RegisterReq)}
}

var _ spec.ServiceDiscovery = (*SimpleSd)(nil)

const logPrefix = "simple_sd"

const (
	httpResOkCode = 200

	registerPath    = "/service/register"
	deregisterPath  = "/service/deregister"
	discoveryPath   = "/service/discovery"
	healthCheckPath = "/service/health_check"
)

func (c *SimpleSd) getRequestUrl(path string) string {
	return fmt.Sprintf("http://localhost:%d%s", c.serverPort, path)
}

func (c *SimpleSd) Name() string {
	return "simple_sd"
}

type httpRes struct {
	Code int // 200 OK
	Msg  string
	Data interface{} `json:"Data,omit_empty"`
}

func (c *SimpleSd) Register(ctx context.Context, service string, host string, port int, metadata map[string]string) error {
	if c.registry[service] != nil {
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
	_, _, errs := gorequest.New().Post(c.getRequestUrl(registerPath)).SendStruct(req).EndStruct(res)
	if len(errs) > 0 {
		return errs[0]
	}
	if res.Code != httpResOkCode {
		return xerr.ErrInternal.New("register failed, got resp: %+v", res)
	}
	c.registry[service] = req
	return nil
}

func (c *SimpleSd) Deregister(ctx context.Context, service string) error {
	c.Lock()
	defer c.Unlock()
	params := c.registry[service]
	if params == nil {
		return xerr.ErrInternal.New("not registered")
	}
	delete(c.registry, params.Id)

	type deregisterReq struct {
		Service string
		Id      string
	}
	req := &deregisterReq{
		Service: service,
		Id:      params.Id,
	}
	res := new(httpRes)
	_, _, errs := gorequest.New().Post(c.getRequestUrl(deregisterPath)).SendStruct(req).EndStruct(res)
	if len(errs) > 0 {
		return errs[0]
	}
	if res.Code != httpResOkCode {
		return xerr.ErrInternal.New("deregister failed, got resp: %+v", res)
	}
	return nil
}

func (c *SimpleSd) Discover(ctx context.Context, serviceName string, block bool) ([]spec.Instance, error) {
	req := &simple_sd.DiscoveryReq{
		Service:   serviceName,
		LastHash:  c.lastHash,
		WaitMaxMs: time.Minute.Milliseconds() * 2,
	}
	if !block {
		req.LastHash = ""
	}
	data := new(simple_sd.DiscoveryRspBody)
	res := &httpRes{Data: data}

	_, _, errs := gorequest.New().Post(c.getRequestUrl(discoveryPath)).SendStruct(req).EndStruct(res)
	if len(errs) > 0 {
		return nil, errs[0]
	}
	if res.Code != httpResOkCode {
		return nil, xerr.ErrInternal.New("discovery failed, got resp: %+v", res)
	}
	c.lastHash = data.Hash
	return lo.Map(data.Instances, func(item simple_sd.ServiceInstance, index int) spec.Instance {
		return spec.Instance{
			ID:       item.Id,
			Name:     item.Name,
			IsUDP:    item.IsUDP,
			Host:     item.Host,
			Port:     item.Port,
			Metadata: item.Metadata,
		}
	}), nil
}

func (c *SimpleSd) HealthCheck(ctx context.Context, service string) error {
	params := c.registry[service]
	if params == nil {
		return xerr.ErrInternal.New("never called Register")
	}
	req := &simple_sd.HealthCheckReq{
		Service: service,
		Id:      params.Id,
	}
	rspBody := new(simple_sd.HealthCheckRspBody)
	res := &httpRes{Data: rspBody}
	_, _, errs := gorequest.New().Post(c.getRequestUrl(healthCheckPath)).SendStruct(req).EndStruct(res)
	if len(errs) > 0 {
		return errs[0]
	}
	if res.Code != httpResOkCode {
		return xerr.ErrInternal.New("health check failed, got resp: %+v", res)
	}
	if !rspBody.Registered {
		xlog.Warn(fmt.Sprintf(logPrefix+": service [%s - id:%s] offline, do re-register now", service, params.Id))
		delete(c.registry, params.Name)
		err := c.Register(ctx, params.Name, params.Host, params.Port, params.Metadata)
		return err
	}
	return nil
}

func (c *SimpleSd) Stop(ctx context.Context) error {
	if c.serverPort == 0 {
		return nil
	}
	c.RLock()
	services := lo.MapToSlice(c.registry, func(item string, _ *simple_sd.RegisterReq) string {
		return item
	})
	c.RUnlock()

	var errs []error
	for _, svc := range services {
		err := c.Deregister(ctx, svc)
		if err != nil {
			errs = append(errs, errors.Wrap(err, fmt.Sprintf("deregister [%s]", svc)))
		}
	}
	return xerr.JoinErrors(errs...)
}
