package etcd

import (
	"context"
	"log"
	"testing"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	etcdEndpoint = "localhost:2379"
)

var (
	testService = "test-service"
	testHost    = "127.0.0.1"
	testPort    = 8080
)

func setupEtcd(t *testing.T) *Etcd {
	etcd, err := New([]string{etcdEndpoint})
	if err != nil {
		t.Fatalf("failed to create etcd client: %v", err)
	}

	// 清理之前的测试数据
	_, err = etcd.cli.Delete(context.Background(), "/services/", clientv3.WithPrefix())
	if err != nil {
		log.Printf("failed to clean up services: %v", err)
	}

	return etcd
}

func TestEtcdIntegration(t *testing.T) {
	etcd := setupEtcd(t)
	defer etcd.Stop(context.TODO())

	ctx := context.Background()

	// 1. Register
	err := etcd.Register(ctx, testService, testHost, testPort, map[string]string{"env": "test"})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 2. Discover
	instances, err := etcd.Discover(ctx, testService, false)
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(instances) == 0 {
		t.Fatalf("Expected at least one instance")
	} else {
		ins := instances[0]
		if ins.Host != testHost || ins.Port != testPort {
			t.Fatalf("Unexpected instance: %+v", ins)
		}
		if ins.Metadata["env"] != "test" {
			t.Fatalf("Metadata not match: %+v", ins.Metadata)
		}
	}

	// 3. HealthCheck
	err = etcd.HealthCheck(ctx, testService)
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}

	ttl, err := etcd.ttl(ctx, testService)
	if err != nil {
		t.Fatalf("ttl failed: %v", err)
	} else if ttl != leaseExpireSeconds-1 {
		t.Fatalf("TTL not right: %d, should be %d", ttl, leaseExpireSeconds-1)
	}

	time.Sleep(time.Second) // after sleep 1s and keep alive once, ttl is still 17=(18-1)

	// 3.1 keepalive
	err = etcd.keepAlive(ctx, testService)
	if err != nil {
		t.Fatalf("keepAlive failed: %v", err)
	}
	ttl, err = etcd.ttl(ctx, testService)
	if err != nil {
		t.Fatalf("ttl failed: %v", err)
	} else if ttl != leaseExpireSeconds-1 {
		t.Fatalf("TTL not right: %d, should be %d", ttl, leaseExpireSeconds-1)
	}

	// 4. Deregister
	err = etcd.Deregister(ctx, testService)
	if err != nil {
		t.Fatalf("Deregister failed: %v", err)
	}

	// 5. Discover again should return empty
	instances, err = etcd.Discover(ctx, testService, false)
	if err != nil {
		t.Fatalf("Discover failed after deregister: %v", err)
	}
	if len(instances) > 0 {
		t.Fatalf("Expected zero instances after deregister")
	}
}
