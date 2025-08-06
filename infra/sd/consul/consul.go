package consul

import (
	"context"
	"fmt"
	capi "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"microsvc/infra/sd/abstract"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/util"
	"microsvc/util/urand"
	"strings"
	"sync"
	"time"
)

type Consul struct {
	client    *capi.Client
	lastIndex uint64
	registry  map[string]*capi.AgentServiceRegistration // svc -> id
	sync.RWMutex
}

var _ abstract.ServiceDiscovery = (*Consul)(nil)
var logPrefix = "consul"

const healthCheckNamePrefix = "microsvc-"

func New(endpoints string) (*Consul, error) {
	cfg := capi.DefaultConfig()
	cfg.Address = endpoints // e.g. 127.0.0.1:8500
	client, err := capi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Consul{client: client, registry: make(map[string]*capi.AgentServiceRegistration)}, nil
}

func (c *Consul) Name() string {
	return "Consul"
}

func (c *Consul) Register(ctx context.Context, serviceName string, host string, port int, metadata map[string]string) error {
	c.Lock()
	defer c.Unlock()
	if c.registry[serviceName] != nil {
		return fmt.Errorf("consul: already registered")
	}
	id := urand.Strings(4)
	tcpAddr := fmt.Sprintf("%s:%d", host, port)
	asr := &capi.AgentServiceRegistration{
		ID:      id,
		Name:    serviceName,
		Tags:    []string{"microsvc"},
		Port:    port,
		Address: host,
		Meta:    metadata,
		Check:   healthCheckAttr(healthCheckNamePrefix+serviceName+"-health", tcpAddr),
	}

	opts := capi.ServiceRegisterOpts{ReplaceExistingChecks: true}.WithContext(ctx)
	err := c.client.Agent().ServiceRegisterOpts(asr, opts)
	if err != nil {
		return err
	}
	c.registry[serviceName] = asr // save register detail that could be used in next registration
	return nil
}

func (c *Consul) Deregister(ctx context.Context, service string) error {
	c.Lock()
	defer c.Unlock()
	if r := c.registry[service]; r == nil {
		return fmt.Errorf("not registered")
	} else {
		delete(c.registry, service)
		opts := (&capi.QueryOptions{}).WithContext(ctx)
		return c.client.Agent().ServiceDeregisterOpts(r.ID, opts)
	}
}

// Discover return a list of instances in healthy status
func (c *Consul) Discover(ctx context.Context, serviceName string, block bool) (list []abstract.Instance, err error) {
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
	c.RLock()
	params := c.registry[service]
	c.RUnlock()
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
		var list []abstract.Instance
		list, err = c.getInstances(service, dur, false)
		lo.ForEach(list, func(item abstract.Instance, index int) {
			if item.ID == params.ID {
				offline = false
			}
		})
	})

	if err != nil {
		return err
	}

	if offline {
		xlog.Warn(fmt.Sprintf(logPrefix+".health-check: service [%s - id:%s] offline, do re-register now", service, params.ID))
		err = c.client.Agent().ServiceRegister(params)
		return err
	}
	return nil
}

// 发现健康的端点列表
func (c *Consul) getInstances(serviceName string, waitTime time.Duration, block bool) (list []abstract.Instance, err error) {
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

		inst := abstract.Instance{
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

func (c *Consul) Stop(ctx context.Context) error {
	if c.client == nil {
		return nil
	}
	c.RLock()
	services := lo.MapToSlice(c.registry, func(item string, _ *capi.AgentServiceRegistration) string {
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
