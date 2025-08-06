//go:build !k8s

package sd

import (
	"context"
	"errors"
	"fmt"
	"microsvc/deploy"
	"microsvc/enums"
	"microsvc/infra/sd/consul"
	"microsvc/infra/sd/etcd"
	"microsvc/infra/sd/mdns"
	"microsvc/infra/sd/simple_sd"
	"microsvc/infra/sd/spec"
	"microsvc/pkg/xlog"
	"microsvc/util"
	"microsvc/util/graceful"
	"microsvc/util/uip"
	simple_sd2 "microsvc/xvendor/simple_sd"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

var registeredServices []string
var gCtx, cancelGCtx = context.WithCancel(context.TODO())

const Impl = "consul" // 统一指定所有服务使用的注册发现组件，支持 consul | etcd | simple_sd | mdns
const logPrefix = "sd: "

var rootSD spec.ServiceDiscovery

func Init(must bool) func(*deploy.XConfig, func(must bool, err error)) {
	return func(cc *deploy.XConfig, finished func(must bool, err error)) {
		graceful.AddStopFunc(Stop)

		var err error

		switch Impl {
		case "simple_sd":
			if cc.SimpleSdHttpPort > 0 {
				if cc.Svc == enums.SvcGateway {
					mustStartSimpleSdServer(cc.SimpleSdHttpPort)
				}
				rootSD = simple_sd.New(cc.SimpleSdHttpPort)
				//tryRunSimpleSdServer(cc.SimpleSdHttpPort)
			} else {
				err = fmt.Errorf("invalid cc.SimpleSdHttpPort: %d", cc.SimpleSdHttpPort)
			}
		case "etcd":
			rootSD, err = etcd.New(cc.ServiceDiscovery.Etcd.Endpoints)
		case "consul":
			rootSD, err = consul.New(strings.Join(cc.ServiceDiscovery.Consul.Endpoints, ","))
		case "mdns": // 仅支持mac
			rootSD = mdns.New()
		default:
			err = fmt.Errorf("invalid sd Impl: %s", Impl)
		}

		if err == nil {
			go startSdDaemon(gCtx)
		}
		finished(must, err)
	}
}

// MustRegister 执行注册服务，失败则panic
func MustRegister(reg ...deploy.RegisterSvc) {
	hasFixedSvcHost := deploy.XConf.ServiceDiscovery.FixedSvcHost != ""
	selfIp := deploy.XConf.ServiceDiscovery.FixedSvcHost
	if selfIp == "" {
		localIps, err := uip.GetLocalPrivateIPs(true, "")
		if err != nil || len(localIps) == 0 {
			xlog.Panic(logPrefix+"GetLocalPrivateIPs failed", zap.Error(err))
		}
		selfIp = localIps[0].String()
	}

	for _, r := range reg {
		name, port := r.RegGRPCBase()
		if name == "" {
			panic(fmt.Sprintf(logPrefix + "service name cannot be empty"))
		}

		util.RunTaskWithCtxTimeout(time.Second*3, func(ctx context.Context) {
			md := r.RegGRPCMeta()
			if hasFixedSvcHost {
				md[spec.MetadataKeyRealHost] = "localhost"
			}
			err := rootSD.Register(ctx, name, selfIp, port, md)
			if err != nil {
				xlog.Panic(logPrefix+"register svc failed", zap.String("sd-name", rootSD.Name()),
					zap.String("reg_svc", name), zap.String("reg_addr", selfIp), zap.Int("port", port), zap.Error(err))
			}
		})
		xlog.Info(logPrefix+"register svc success", zap.String("sd-name", rootSD.Name()),
			zap.String("reg_svc", name),
			zap.String("addr", fmt.Sprintf("%s:%d", selfIp, port)))

		registeredServices = append(registeredServices, name)
	}
}

func Stop() {
	cancelGCtx()

	var err error
	isTimeout := util.RunTaskWithCtxTimeout(time.Second*10, func(ctx context.Context) {
		err = rootSD.Stop(ctx)
	})
	if isTimeout || err != nil {
		xlog.Error(fmt.Sprintf("sd: [%s] resource release failed", Impl), zap.Error(err), zap.Bool("timeout", isTimeout))
		return
	}
	xlog.Debug(fmt.Sprintf("sd: [%s] resource released...", Impl))
}

// startSdDaemon automatically reconnect the service to the registry center in case of service
// unregister due to registry center abnormalities.
func startSdDaemon(ctx context.Context) {
	var err error
	var errCnt int
	var ticker = time.NewTicker(spec.HealthCheckInterval)
	for {
		select {
		case <-ticker.C: // health checking
			for _, service := range registeredServices {
				util.RunTaskWithCtxTimeout(time.Second*3, func(ctx context.Context) {
					err = rootSD.HealthCheck(ctx, service)
					if err != nil {
						xlog.Error("sd-daemon: HealthCheck failed", zap.String("service", service), zap.Error(err), zap.Int("errCnt", errCnt))
					}
				})
			}
		case <-ctx.Done():
			return
		}
	}
}

func tryRunSimpleSdServer(port int) {
	server := simple_sd2.NewSimpleSdHTTPServer(port)

	// 修改为DEBUG 可进行调试
	simple_sd2.SetLogLevel(simple_sd2.LogLevelInfo)
	//simple_sd2.SetLogLevel(simple_sd2.LogLevelDebug)

	if server.IsRunningOnLocalHost() {
		xlog.Debug(logPrefix + fmt.Sprintf("simple_sd server is already running on local:%d", port))
		return
	}
	xlog.Debug(logPrefix + "no simple_sd server found, start it on localhost:" + fmt.Sprintf("%d", port))

	go func() {
		err := server.Run()
		if !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	time.Sleep(time.Millisecond * 500)
	if !server.IsRunningOnLocalHost() {
		panic("SimpleSd server start failed")
	}
}

func mustStartSimpleSdServer(port int) {
	server := simple_sd2.NewSimpleSdHTTPServer(port)

	// 修改为DEBUG 可进行调试
	simple_sd2.SetLogLevel(simple_sd2.LogLevelInfo)
	//simple_sd2.SetLogLevel(simple_sd2.LogLevelDebug)

	xlog.Debug(logPrefix + "start it on localhost:" + fmt.Sprintf("%d", port))

	go func() {
		err := server.Run()
		if !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	time.Sleep(time.Millisecond * 500)
	if !server.IsRunningOnLocalHost() {
		panic("SimpleSd server start failed")
	}
}
