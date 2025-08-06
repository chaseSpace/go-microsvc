package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/k0kubun/pp/v3"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	clientv3 "go.etcd.io/etcd/client/v3"
	"microsvc/infra/sd/spec"
	"microsvc/pkg/xerr"
	"microsvc/util/urand"
	"sync"
	"time"
)

const (
	leaseExpireSeconds = 18

	keyPrefix = "/services/"
)

type Etcd struct {
	cli      *clientv3.Client
	registry map[string]clientv3.LeaseID
	sync.RWMutex
}

var _ spec.ServiceDiscovery = (*Etcd)(nil)

const logPrefix = "etcd"

// New 创建etcd注册发现客户端
// 简介：etcd 是一个 分布式、高可用、强一致性的键值存储系统
func New(endpoints []string) (*Etcd, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 3 * time.Second,
	})
	return &Etcd{cli: cli, registry: make(map[string]clientv3.LeaseID)}, err
}

func (c *Etcd) Name() string {
	return "etcd"
}

func (c *Etcd) Register(ctx context.Context, service string, host string, port int, metadata map[string]string) error {
	ctx = clientv3.WithRequireLeader(ctx) // 确保集群有leader时返回结果，否则返回Err
	// 1. 创建租约
	lease := clientv3.NewLease(c.cli)
	leaseResp, err := lease.Grant(ctx, leaseExpireSeconds)
	if err != nil {
		return err
	}

	c.Lock()
	c.registry[service] = leaseResp.ID
	c.Unlock()

	id := urand.Strings(4)

	// 2. 写入 kv（带租约）
	key := fmt.Sprintf(keyPrefix+"%s/%s", service, id)
	val, _ := json.Marshal(&spec.Instance{
		ID:       id,
		Name:     service,
		IsUDP:    false,
		Host:     host,
		Port:     port,
		Metadata: metadata,
	})

	_, err = c.cli.Put(ctx, key, string(val), clientv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	return err
}

func (c *Etcd) Deregister(ctx context.Context, service string) (err error) {
	ctx = clientv3.WithRequireLeader(ctx) // 确保集群有leader时返回结果，否则返回Err

	c.Lock()
	defer c.Unlock()
	leaseID := c.registry[service]
	delete(c.registry, service)

	if leaseID == 0 {
		return fmt.Errorf("not registered")
	}
	lease := clientv3.NewLease(c.cli)
	_, err = lease.Revoke(ctx, leaseID)
	return
}

func (c *Etcd) Discover(ctx context.Context, service string, block bool) (list []spec.Instance, err error) {
	ctx = clientv3.WithRequireLeader(ctx) // 确保集群有leader时返回结果，否则返回Err

	resp, err := c.cli.Get(ctx, fmt.Sprintf(keyPrefix+"%s/", service), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	rmap := map[string]spec.Instance{}
	defer func() {
		for _, ins := range rmap {
			list = append(list, ins)
		}
	}()

	for _, kv := range resp.Kvs {
		var ins spec.Instance
		if err := json.Unmarshal(kv.Value, &ins); err == nil {
			rmap[string(kv.Key)] = ins
		} else {
			return nil, fmt.Errorf("json.Unmarshal error: %w", err)
		}
	}
	if !block {
		return
	}

	// Block=true: 使用watch api
	watch := clientv3.NewWatcher(c.cli)
	watchChan := watch.Watch(ctx, fmt.Sprintf(keyPrefix+"%s/", service), clientv3.WithPrefix())
	//_, _ = pp.Printf("svccli: [%s] is watching service %s...\n", "etcd", service)

	for {
		select {
		case <-ctx.Done():
			return
		case wres := <-watchChan:
			for _, ev := range wres.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					var ins spec.Instance
					if err := json.Unmarshal(ev.Kv.Value, &ins); err == nil {
						rmap[string(ev.Kv.Key)] = ins
					} else {
						return nil, fmt.Errorf("json.Unmarshal error: %w", err)
					}
					_, _ = pp.Printf("svccli: [%s] service UP: %s\n", "etcd", string(ev.Kv.Key))
				case clientv3.EventTypeDelete:
					delete(rmap, string(ev.Kv.Key))
					_, _ = pp.Printf("svccli: [%s] service DOWN: %s\n", "etcd", string(ev.Kv.Key))
				}
			}
			return
		}
	}
}

func (c *Etcd) HealthCheck(ctx context.Context, service string) error {
	c.RLock()
	leaseID := c.registry[service]
	c.RUnlock()

	if leaseID == 0 {
		return fmt.Errorf("no active lease")
	}

	resp, err := c.cli.TimeToLive(ctx, leaseID)
	if err != nil {
		return err
	}

	if resp.TTL <= leaseExpireSeconds/2 { // 续约
		//_, _ = pp.Printf("%s sd: [%s] lease is approaching expiration in %vs, renewing..\n", time.Now().Format(time.DateTime), "etcd", resp.TTL)
		err = c.keepAlive(ctx, service)
	}

	return err
}

func (c *Etcd) keepAlive(ctx context.Context, service string) error {
	c.RLock()
	leaseID := c.registry[service]
	c.RUnlock()

	_, err := c.cli.KeepAliveOnce(ctx, leaseID)
	return err
}

func (c *Etcd) ttl(ctx context.Context, service string) (int64, error) {
	c.RLock()
	leaseID := c.registry[service]
	c.RUnlock()

	r, err := c.cli.TimeToLive(ctx, leaseID)
	if err != nil {
		return 0, err
	}
	return r.TTL, nil
}

func (c *Etcd) Stop(ctx context.Context) error {
	if c.cli == nil {
		return nil
	}
	c.RLock()
	services := lo.MapToSlice(c.registry, func(item string, _ clientv3.LeaseID) string {
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
