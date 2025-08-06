package abstract

import (
	"context"
	"time"
)

type Empty struct {
}

var _ ServiceDiscovery = new(Empty)

func (e Empty) Name() string {
	return "empty"
}

func (e Empty) Register(ctx context.Context, serviceName string, address string, port int, metadata map[string]string) error {
	return nil
}

func (e Empty) Deregister(ctx context.Context, serviceName string) error {
	return nil
}

func (e Empty) Discover(ctx context.Context, serviceName string, block bool) ([]Instance, error) {
	if block {
		time.Sleep(time.Millisecond * 100) // mock block
	}
	return nil, nil
}

func (e Empty) HealthCheck(ctx context.Context, service string) error {
	return nil
}

func (e Empty) Stop(ctx context.Context) error { return nil }
