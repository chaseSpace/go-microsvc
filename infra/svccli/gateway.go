//go:build !k8s

package svccli

import (
	"microsvc/enums"
	"microsvc/infra/sd"
	"microsvc/infra/xgrpc"
	"sync"

	"google.golang.org/grpc"
)

type InstanceMgr struct {
	cmap sync.Map
	mu   sync.RWMutex
}

type InstanceImplT struct {
	impl *sd.InstanceImpl
	once *sync.Once
}

var InstMgr = &InstanceMgr{}

func GetConn(svc enums.Svc) (conn *grpc.ClientConn) {
	var inst *InstanceImplT

	defer func() {
		if cw, err := inst.impl.GetSingleConnWrapper(); err == nil {
			conn = cw.Conn
			return
		}
		conn = xgrpc.NewInvalidGRPCConn(svc.Name())
	}()

	// 1. quick path
	// - sync.Map 支持无锁并发
	v, ok := InstMgr.cmap.Load(svc)
	if ok {
		inst = v.(*InstanceImplT)
		return
	}

	// 2. slow path
	v, _ = InstMgr.cmap.LoadOrStore(svc, &InstanceImplT{
		impl: sd.NewInstance(svc.Name(), func(conn *grpc.ClientConn) interface{} {
			return nil
		}, rootSD),
		once: new(sync.Once),
	})
	inst = v.(*InstanceImplT)
	inst.once.Do(func() { // 每个svc只初始化一个instance
		inst.impl = sd.NewInstance(svc.Name(),
			func(conn *grpc.ClientConn) interface{} { return nil },
			rootSD)
	})
	return
}
