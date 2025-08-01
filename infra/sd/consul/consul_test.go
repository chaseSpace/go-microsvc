package consul

import (
	"context"
	"testing"
	"time"
)

func TestConsulSD(t *testing.T) {
	sd, err := New("127.0.0.1:8500")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { sd.Stop() })

	serviceName := "consul-svc5"
	host := "127.0.0.1"
	port := 8500
	t.Run("Register", func(t *testing.T) {
		if err := sd.Register(serviceName, host, port, nil); err != nil {
			t.Fatalf("Register: %v", err)
		}
	})

	time.Sleep(time.Second * 5) // 等待consul检测状态，才能Discovery
	t.Run("DiscoverAfterRegister", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
		defer cancel()

		list, err := sd.Discover(ctx, serviceName, false)
		if err != nil {
			t.Fatalf("Discover: %v", err)
		}
		if len(list) == 0 {
			t.Fatal("expected at least 1 instance")
		}
		t.Logf("instances: %+v", list)
	})

	t.Run("Deregister", func(t *testing.T) {
		if err := sd.Deregister(serviceName); err != nil {
			t.Fatalf("Deregister: %v", err)
		}
	})

	t.Run("DiscoverAfterDeregister", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
		defer cancel()

		list, err := sd.Discover(ctx, serviceName, false)
		if err != nil {
			t.Fatalf("Discover: %v", err)
		}
		if len(list) != 0 {
			t.Fatalf("expected 0 instances, got %d", len(list))
		}
	})
}
