//go:build !k8s

package svccli

import (
	"fmt"
	"microsvc/deploy"
	"microsvc/enums"
	"microsvc/infra/sd"
	"microsvc/infra/sd/abstract"
	"microsvc/infra/sd/consul"
	"microsvc/infra/sd/mdns"
	"microsvc/infra/sd/simple_sd"
	"microsvc/infra/xgrpc"
	"microsvc/pkg/xlog"
	"microsvc/util/graceful"
	"sync"
)

/*
如果使用DNS名称连接服务，则不需要调用Init函数
*/

const impl = sd.Impl

var rootSD abstract.ServiceDiscovery

func SetDefaultSD(sd abstract.ServiceDiscovery) {
	rootSD = sd
}

func Init(must bool) func(*deploy.XConfig, func(must bool, err error)) {
	graceful.AddStopFunc(Stop)

	return func(cc *deploy.XConfig, onEnd func(must bool, err error)) {
		var err error

		switch impl {
		case "simple_sd":
			rootSD = simple_sd.New(cc.SimpleSdHttpPort)
		case "consul":
			rootSD, err = consul.New(cc.ServiceDiscovery.Consul.Address)
		case "mdns":
			rootSD = mdns.New()
		default:
			err = fmt.Errorf("invalid sd impl: %s", impl)
		}

		onEnd(must, err)
	}
}

type RpcClient struct {
	once      sync.Once
	svc       enums.Svc
	inst      *sd.InstanceImpl
	genClient sd.GenClient
}

func NewCli(svc enums.Svc, gc sd.GenClient) *RpcClient {
	cli := &RpcClient{svc: svc, genClient: gc}
	return cli
}

// Getter returns gRPC Server Client
func (c *RpcClient) Getter() any {
	c.once.Do(func() {
		c.inst = sd.NewInstance(c.svc.Name(), c.genClient, rootSD)
		initializedSvcCli = append(initializedSvcCli, c)
	})
	v, err := c.inst.GetSingleConnWrapper()
	if err == nil {
		return v.RpcClient
	}
	return c.genClient(xgrpc.NewInvalidGRPCConn(c.svc.Name()))
}

func (c *RpcClient) Stop() {
	if c.inst != nil {
		c.inst.Stop()
	}
}

var initializedSvcCli []*RpcClient

func Stop() {
	for _, svcCli := range initializedSvcCli {
		svcCli.Stop()
	}
	xlog.Debug("svccli: resource released...")
}
