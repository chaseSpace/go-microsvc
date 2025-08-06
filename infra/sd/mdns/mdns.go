package mdns

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
	"microsvc/infra/sd/spec"
	"microsvc/pkg/xerr"
	"microsvc/util/urand"
	"microsvc/xvendor/mdns"
	"net"
	"strings"
	"sync"
	"time"
)

/*
简介:
 	mDNS（Multicast DNS，多播 DNS）是一种零配置网络协议（Zeroconf），
	用于在本地局域网（如家庭或公司内网）中实现无需 DNS 服务器即可进行主机名解析的功能。

警告：
	mDNS 不适合用于生产环境，它仅适用于本地局域网使用。
	此外，在 Windows 系统中，mDNS 的支持可能不稳定。如要在 Windows 体验完整支持 mDNS 功能，
	建议安装：Bonjour Print Services for Windows
*/

// Mdns implements the spec.ServiceDiscovery with mDNS (Multicast DNS)
// protocol using UDP.
type Mdns struct {
	registry map[string]*registryStub // svc -> id
	sync.RWMutex
	instCache map[string][]string
}

type registryStub struct {
	s  *mdns.Server
	id string
}

var _ spec.ServiceDiscovery = (*Mdns)(nil)

const logPrefix = "mdns"

// mDnsDomain 是 mDNS 使用的域名后缀（如 .local）。设置该域名是必要的，
// 因为 mDNS 是一种多播协议，通过域名来过滤和匹配局域网中的服务公告包。
const mDnsDomain = "microsvc."

func New() *Mdns {
	return &Mdns{registry: make(map[string]*registryStub), instCache: map[string][]string{}}
}

func (c *Mdns) Name() string {
	return "mDNS"
}

func getServerId(completeName string) string {
	name := strings.Split(completeName, ".")[0]
	return strings.TrimSpace(name)
}

// Register register a service instance
// NOTE: 三方库不支持 metadata
func (c *Mdns) Register(ctx context.Context, serviceName string, host string, port int, metadata map[string]string) (err error) {
	c.Lock()
	defer c.Unlock()
	if c.registry[serviceName] != nil {
		return fmt.Errorf("already registered")
	}
	id := urand.Strings(4)

	ds, err := mdns.NewMDNSService(id, serviceName, mDnsDomain, "", port, []net.IP{net.ParseIP(host)}, nil)
	if err != nil {
		return err
	}
	server, err := mdns.NewServer(&mdns.Config{Zone: ds})
	if err != nil {
		return err
	}
	c.registry[serviceName] = &registryStub{
		s:  server,
		id: id,
	}
	return
}

func (c *Mdns) Deregister(ctx context.Context, service string) error {
	c.Lock()
	defer c.Unlock()
	if rs := c.registry[service]; rs == nil {
		return fmt.Errorf("not registered")
	} else {
		delete(c.registry, service)
		return rs.s.Shutdown()
	}
}

func (c *Mdns) Discover(ctx context.Context, svc string, block bool) (instances []spec.Instance, err error) {
	if !block {
		return c.discoverOnce(ctx, svc)
	}
	return c.discoverWithWatch(ctx, svc)
}

func (c *Mdns) discoverOnce(ctx context.Context, svc string) (instances []spec.Instance, err error) {
	var result []spec.Instance
	entries := make(chan *mdns.ServiceEntry, 10)

	done := make(chan bool, 1)
	go func() {
		defer close(done)
		for entry := range entries {
			result = append(result, c.entryToInstance(entry))
		}
	}()

	err = mdns.Lookup(svc, mDnsDomain, time.Second, entries)
	close(entries)
	if err != nil {
		return nil, err
	}

	<-done
	c.updateCacheAndIsChanged(svc, result)
	return result, nil
}

func (c *Mdns) discoverWithWatch(ctx context.Context, svc string) (instances []spec.Instance, err error) {
	// 由于mdns协议没有server端，所以只能通过高频轮询来获得更高的实时性（服务上下线的感知）
	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()
	var result []spec.Instance

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			result = nil
			entries := make(chan *mdns.ServiceEntry, 10)

			done := make(chan bool)
			go func() {
				defer close(done)
				for entry := range entries {
					result = append(result, c.entryToInstance(entry))
				}
			}()

			err = mdns.Lookup(svc, mDnsDomain, time.Second, entries)
			close(entries)
			if err != nil {
				continue
			}

			<-done

			if c.updateCacheAndIsChanged(svc, result) {
				return result, nil
			}
		}
	}
}

func (c *Mdns) HealthCheck(ctx context.Context, service string) error {
	c.RLock()
	defer c.RUnlock()
	if c.registry[service] == nil {
		return fmt.Errorf("not registered")
	}
	// NOTE: not need
	return nil
}

func (c *Mdns) entryToInstance(entry *mdns.ServiceEntry) spec.Instance {
	return spec.Instance{
		ID:       getServerId(entry.Name),
		Name:     entry.Name,
		Host:     entry.AddrV4.String(),
		Port:     entry.Port,
		Metadata: nil,
	}
}

func (c *Mdns) updateCacheAndIsChanged(serviceName string, newInstances spec.InstanceSlice) bool {
	c.Lock()
	defer c.Unlock()
	cachedIds := c.instCache[serviceName]
	ids := newInstances.SortedIds()
	c.instCache[serviceName] = ids
	return !slices.Equal(ids, cachedIds)
}

func (c *Mdns) Stop(ctx context.Context) error {
	c.RLock()
	services := lo.MapToSlice(c.registry, func(item string, value *registryStub) string {
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
