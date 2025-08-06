package abstract

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"sort"
)

type ServiceDiscovery interface {
	Name() string
	Register(ctx context.Context, service string, host string, port int, metadata map[string]string) error
	Deregister(ctx context.Context, service string) error
	Discover(ctx context.Context, service string, block bool) ([]Instance, error)
	HealthCheck(ctx context.Context, service string) error // 包含保活的逻辑
	Stop(ctx context.Context) error
}

// Instance 表示注册的单个实例
type Instance struct {
	ID       string
	Name     string
	IsUDP    bool
	Host     string
	Port     int
	Metadata map[string]string
}

func (s Instance) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type InstanceSlice []Instance

func (slice InstanceSlice) SortedIds() (list []string) {
	lo.ForEach(slice, func(item Instance, index int) {
		list = append(list, item.ID)
	})
	sort.Strings(list)
	return
}

type CtxDurKey struct{}
