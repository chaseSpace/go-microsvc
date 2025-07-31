//go:build !k8s

package sd

import (
	"context"
	"errors"
	"fmt"
	"microsvc/deploy"
	"microsvc/infra/sd/abstract"
	"microsvc/infra/sd/simple_sd"
	"microsvc/pkg/xlog"
	"microsvc/util"
	"microsvc/util/graceful"
	"microsvc/util/uip"
	simple_sd2 "microsvc/xvendor/simple_sd"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var registeredServices []string
var gCtx, cancelGCtx = context.WithCancel(context.TODO())

const logPrefix = "sd: "

var rootSD abstract.ServiceDiscovery

func Init(must bool) func(*deploy.XConfig, func(must bool, err error)) {
	graceful.AddStopFunc(Stop)

	return func(cc *deploy.XConfig, finished func(must bool, err error)) {
		var err error
		if cc.SimpleSdHttpPort > 0 {
			rootSD = simple_sd.New(cc.SimpleSdHttpPort)
			//tryRunSimpleSdServer(cc.SimpleSdHttpPort)
			go startSdDaemon(gCtx)
		} else {
			err = fmt.Errorf("invalid cc.SimpleSdHttpPort: %d", cc.SimpleSdHttpPort)
		}

		// take consul or etcd(not have yet) in your like
		//rootSD, err = consul.New()
		//if err != nil {
		//	xlog.Error(logPrefix+"New failed", zap.Error(err))
		//}
		finished(must, err)
	}
}

func InitSimpleSdServer() func(*deploy.XConfig, func(must bool, err error)) {
	return func(cc *deploy.XConfig, finished func(must bool, err error)) {
		mustStartSimpleSdServer(cc.SimpleSdHttpPort)
	}
}

// MustRegister 执行注册服务，失败则panic
// 如果使用DNS名称链接服务，则不需要注册
func MustRegister(reg ...deploy.RegisterSvc) {
	selfIp := "127.0.0.1"
	if !deploy.XConf.Env.IsDev() {
		localIps, err := uip.GetLocalPrivateIPs(true, "")
		if err != nil || len(localIps) == 0 {
			xlog.Panic(logPrefix+"GetLocalPrivateIPs failed", zap.Error(err))
		}
		selfIp = localIps[0].String()
	}

	for _, r := range reg {
		name, addr, port := r.RegGRPCBase()
		if name == "" {
			panic(fmt.Sprintf(logPrefix + "service name cannot be empty"))
		}
		if addr == "" {
			addr = selfIp
		}
		err := rootSD.Register(name, addr, port, r.RegGRPCMeta())
		if err != nil {
			xlog.Panic(logPrefix+"register svc failed", zap.String("sd-name", rootSD.Name()),
				zap.String("reg_svc", name), zap.String("reg_addr", addr), zap.Int("port", port), zap.Error(err))
		}
		xlog.Info(logPrefix+"register svc success", zap.String("sd-name", rootSD.Name()),
			zap.String("reg_svc", name),
			zap.String("addr", fmt.Sprintf("%s:%d", addr, port)))

		registeredServices = append(registeredServices, name)
	}
}

func Stop() {
	cancelGCtx()
	for _, s := range registeredServices {
		err := rootSD.Deregister(s)
		if err != nil {
			xlog.Error(logPrefix+"deregister fail", zap.String("sd-name", rootSD.Name()), zap.Error(err), zap.String("svc", s))
		} else {
			xlog.Info(logPrefix+"deregister success", zap.String("sd-name", rootSD.Name()), zap.String("svc", s))
		}
	}
}

// startSdDaemon automatically reconnect the service to the registry center in case of service
// unregister due to registry center abnormalities.
func startSdDaemon(ctx context.Context) {
	var err error
	var errCnt int
	var ticker = time.NewTicker(time.Second * 5)
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
