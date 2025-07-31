package sd

import (
	"container/list"
	"context"
	"errors"
	"microsvc/infra/sd/abstract"
	"microsvc/infra/xgrpc"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type InstanceImpl struct {
	svc string

	entryStore sync.Map

	// A linked list is used to store all gRPC-client object that belong to a
	// single service, ensuring efficient management and access.
	// The maximum length of linked-list is determined by the number of target service instances.
	grpcConns *list.List
	curr      *list.Element // current element
	quit      chan struct{}
	genClient GenClient
	sd        abstract.ServiceDiscovery
}

type GrpcConnWrapper struct {
	addr string
	// One Conn represents a grpc-client which correspond to target endpoint.
	// Conn could be repeatedly use.
	Conn      *grpc.ClientConn
	RpcClient interface{}
}

type GenClient func(conn *grpc.ClientConn) interface{}

func NewInstance(svc string, genClient GenClient, discovery abstract.ServiceDiscovery) *InstanceImpl {
	ins := &InstanceImpl{
		svc:        svc,
		entryStore: sync.Map{},
		grpcConns:  list.New(),
		genClient:  genClient,
		quit:       make(chan struct{}),
		sd:         discovery,
	}
	_ = ins.query(false) // first time query to respond to first rpc call
	go ins.backgroundRefresh()
	return ins
}

// GetSingleConnWrapper get next conn, here implement load balancing（svc node）
func (i *InstanceImpl) GetSingleConnWrapper() (instance *GrpcConnWrapper, err error) {
	instance, err = i.getCurrConn()
	if err != nil || instance != nil {
		return
	}
	// linked list is empty, try to refresh without blocking
	_ = i.query(false)
	return i.getCurrConn()
}

func (i *InstanceImpl) getCurrConn() (instance *GrpcConnWrapper, err error) {
	for i.curr != nil {
		instance = i.curr.Value.(*GrpcConnWrapper)
		// then we move the curr ptr to next or first element
		if next := i.curr.Next(); next != nil {
			i.curr = next
		} else {
			i.curr = i.grpcConns.Front()
		}
		if i.isConnReady(instance) {
			return
		}
	}
	return nil, xerr.ErrServiceUnavailable.AppendMsg(i.svc)
}

func (i *InstanceImpl) isConnReady(instance *GrpcConnWrapper) bool {
	// if conn state is idle, do connect
	if instance.Conn.GetState() == connectivity.Idle {
		instance.Conn.Connect()
		return true
	}
	// if the conn is shutdown(contains `closing`), remove it then try next one in outside.
	if instance.Conn.GetState() == connectivity.Shutdown {
		i.entryStore.Delete(instance.addr)
		i.removeInstance(instance.addr)
		return false
	}
	return true
}

func (i *InstanceImpl) backgroundRefresh() {
	block := true
	for {
		st := time.Now()
		err := i.query(block)
		select {
		case <-i.quit:
			xlog.Debug(logPrefix+"quited", zap.String("Svc", i.svc))
			return
		default:
			if err != nil {
				block = false // To fix sd client can't refresh svc entries after sd server restarted.
				xlog.Error(logPrefix+"query err, hold on...", zap.Error(err), zap.String("target", i.svc))
				time.Sleep(time.Second * 3)
			} else {
				block = true
				if time.Since(st) < time.Second {
					xlog.Warn(logPrefix+"query too fast, hold on...", zap.String("target", i.svc))
					time.Sleep(time.Second * 3)
				}
			}
		}
	}
}

// 阻塞刷新（首次请求不阻塞）
func (i *InstanceImpl) query(block bool) error {
	var (
		entries []abstract.ServiceInstance
		cc      *grpc.ClientConn
		err     error
		ctx     context.Context
	)
	discovery := func() ([]abstract.ServiceInstance, error) {
		ctx = context.WithValue(context.Background(), abstract.CtxDurKey{}, time.Minute*2)
		return i.sd.Discover(ctx, i.svc, block)
	}

	//println("query-111111", i.svc)
	//defer func() {
	//	println("query-222222", i.svc)
	//}()
	entries, err = discovery() // simple-sd server down!!!
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			xlog.Debug(logPrefix+"discover timeout", zap.String("svc", i.svc))
		} else {
			xlog.Error(logPrefix+"discover fail", zap.Error(err), zap.String("svc", i.svc))
		}
		return err
	}

	if len(entries) == 0 {
		xlog.Warn(logPrefix+"discover nothing", zap.String("service", i.svc))
	} else {
		xlog.Debug(logPrefix+"discover result", zap.Any("entries", entries))
	}

	var availableEntries = make(map[string]int8)
	for _, entry := range entries {
		addr := entry.Addr()
		availableEntries[addr] = 1
		if _, ok := i.entryStore.Load(addr); ok {
			continue
		}
		// Add new grpc conn
		cc, err = xgrpc.NewGRPCClient(addr, i.svc)
		if err == nil {
			xlog.Debug(logPrefix+"newGRPCClient OK", zap.String("addr", addr))
			connWrapper := &GrpcConnWrapper{addr: addr, RpcClient: i.genClient(cc), Conn: cc}
			i.entryStore.Store(addr, connWrapper)
			i.grpcConns.PushBack(connWrapper)
			if i.curr == nil {
				i.curr = i.grpcConns.Front()
			}
		} else {
			xlog.Error(logPrefix+"newGRPCClient failed", zap.Error(err))
		}
	}

	// Clean unavailable entries
	i.entryStore.Range(func(key, value interface{}) bool {
		addr := key.(string)
		conn := value.(*GrpcConnWrapper)
		if availableEntries[addr] == 0 {
			_ = conn.Conn.Close()
			i.entryStore.Delete(addr)
			i.removeInstance(addr)
			xlog.Debug(logPrefix+"removeInstance", zap.String("addr", addr))
		}
		return true
	})

	return nil
}

func (i *InstanceImpl) Stop() {
	close(i.quit)
}

func (i *InstanceImpl) removeInstance(addr string) {
	curr := i.grpcConns.Front()
	for curr != nil {
		if curr.Value.(*GrpcConnWrapper).addr == addr {
			i.grpcConns.Remove(curr)
			return
		}
		curr = curr.Next()
	}
}
