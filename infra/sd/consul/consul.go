package consul

import (
	"context"
	"fmt"
	capi "github.com/hashicorp/consul/api"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"microsvc/infra/sd/abstract"
	"microsvc/pkg/xlog"
	"microsvc/util"
	"microsvc/util/urand"
	"strings"
	"time"
)

type Consul struct {
	client    *capi.Client
	lastIndex uint64
	registry  map[string]*capi.AgentServiceRegistration // svc -> id
}

var _ abstract.ServiceDiscovery = (*Consul)(nil)
var logPrefix = "consul."

const Name = "Consul"
const healthCheckNamePrefix = "microsvc-"

func New(address string) (*Consul, error) {
	cfg := capi.DefaultConfig()
	cfg.Address = address // e.g. 127.0.0.1:8500
	client, err := capi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Consul{client: client, registry: make(map[string]*capi.AgentServiceRegistration)}, nil
}

func (c *Consul) Name() string {
	return Name
}

func (c *Consul) Register(serviceName string, host string, port int, metadata map[string]string) error {
	if c.registry[serviceName] != nil {
		return fmt.Errorf("consul: already registered")
	}
	id := urand.Strings(4)
	tcpAddr := fmt.Sprintf("%s:%d", host, port)
	params := &capi.AgentServiceRegistration{
		ID:      id,
		Name:    serviceName,
		Tags:    []string{"microsvc"},
		Port:    port,
		Address: host,
		Meta:    metadata,
		Check:   healthCheckAttr(healthCheckNamePrefix+serviceName+"-health", tcpAddr),
	}

	err := c.client.Agent().ServiceRegister(params)
	if err != nil {
		return err
	}
	c.registry[serviceName] = params // save register params that could be used in next registration
	return nil
}

func (c *Consul) Deregister(service string) error {
	if r := c.registry[service]; r == nil {
		return fmt.Errorf("not registered")
	} else {
		delete(c.registry, service)
		return c.client.Agent().ServiceDeregister(r.ID)
	}
}

// Discover return a list of instances in healthy status
func (c *Consul) Discover(ctx context.Context, serviceName string, block bool) (list []abstract.ServiceInstance, err error) {
	err = context.DeadlineExceeded // default
	dur := time.Minute
	if val := ctx.Value(abstract.CtxDurKey{}); val != nil {
		dur = val.(time.Duration) // use duration here, because Consul do not support block by context
	}
	util.RunTask(ctx, func() {
		list, err = c.getInstances(serviceName, dur, block)
	})
	return
}

func (c *Consul) HealthCheck(ctx context.Context, service string) error {
	params := c.registry[service]
	if params == nil {
		return fmt.Errorf("not registered")
	}
	err := context.DeadlineExceeded // default
	dur := time.Minute
	if val := ctx.Value(abstract.CtxDurKey{}); val != nil {
		dur = val.(time.Duration) // use duration here, because Consul do not support block by context
	}

	offline := true
	util.RunTask(ctx, func() {
		var list []abstract.ServiceInstance
		list, err = c.getInstances(service, dur, false)
		lo.ForEach(list, func(item abstract.ServiceInstance, index int) {
			if item.ID == params.ID {
				offline = false
			}
		})
	})

	if err != nil {
		return err
	}

	if offline {
		xlog.Warn(fmt.Sprintf(logPrefix+"HealthCheck: service [%s - id:%s] offline, do re-register now", service, params.ID))
		err = c.client.Agent().ServiceRegister(params)
		return err
	}
	return nil
}

// 发现健康的端点列表
func (c *Consul) getInstances(serviceName string, waitTime time.Duration, block bool) (list []abstract.ServiceInstance, err error) {
	opt := &capi.QueryOptions{WaitIndex: c.lastIndex, WaitTime: waitTime, UseCache: true, MaxAge: time.Minute * 5}
	if !block {
		opt.WaitIndex = 0 // set to 0 to disable blocking query
	}
	// 即使这里指定了 passingOnly=true，api仍然会返回 Service check fail的端点，下面for循环中会进行二次过滤
	entries, meta, err := c.client.Health().Service(serviceName, "", true, opt)
	if err != nil {
		return nil, err
	}
	if c.lastIndex > meta.LastIndex { //  index goes backwards, reset it
		c.lastIndex = 0
	} else if c.lastIndex < meta.LastIndex {
		c.lastIndex = meta.LastIndex
	}

	var checkPass bool
	for _, s := range entries {

		checkPass = false
		for _, check := range s.Checks {
			if strings.HasPrefix(check.CheckID, healthCheckNamePrefix) && check.Status == "passing" {
				checkPass = true
			}
		}
		if !checkPass {
			continue
		}

		inst := abstract.ServiceInstance{
			ID:       s.Service.ID,
			Name:     serviceName,
			Host:     s.Service.Address,
			Port:     s.Service.Port,
			Metadata: s.Service.Meta,
		}
		list = append(list, inst)
	}
	return list, nil
}

func (c *Consul) Stop() {
	for _, r := range c.registry {
		err := c.client.Agent().ServiceDeregister(r.ID)
		if err != nil {
			xlog.Error(logPrefix+"Stop: deregister fail", zap.Error(err), zap.String("svc", r.Name))
		} else {
			xlog.Info(logPrefix+"Stop: deregister success", zap.String("svc", r.Name))
		}
	}
}
