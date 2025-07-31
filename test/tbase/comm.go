package tbase

import (
	"context"
	"microsvc/consts"
	"microsvc/deploy"
	"microsvc/enums"
	"microsvc/infra"
	"microsvc/infra/sd/abstract"
	"microsvc/infra/svccli"
	"microsvc/infra/xgrpc"
	"microsvc/pkg"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/adminpb"
	commonpb "microsvc/protocol/svc/commonpb"
	"microsvc/util"
	"microsvc/util/graceful"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cast"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

var oncemap sync.Map

func init() {
	_ = os.Setenv(consts.EnvVarLogLevel, "debug")
	_ = os.Setenv(consts.EnvNoPrintCfg, "1")
}

// TearUp 这个方法只完成grpc服务的client初始化，所以你需要提前在本地启动待测试的服务
func TearUp(svc enums.Svc, svcConf deploy.SvcConfImpl) {
	var o = new(sync.Once)
	v, ok := oncemap.Load(svc.Name())
	if !ok {
		oncemap.Store(svc.Name(), o)
	} else {
		o = v.(*sync.Once)
	}
	o.Do(func() {
		//println(111, svc.Name())
		graceful.SetupSignal()
		if !isProjectRootDir() {
			wd, _ := os.Getwd()
			parentDir := filepath.Dir(filepath.Dir(wd))
			_ = os.Chdir(parentDir)
		}
		_ = os.Setenv(consts.EnvNoPrintCfg, "1")
		deploy.InitCfg(svc, svcConf)
		pkg.Setup(
			xlog.Init,
		)
		infra.Setup(
			//sd.Init(true),
			svccli.Init(true),
		)
	})
}

func TearDown() {
	graceful.OnExit()
}

var oncemapEmptySD sync.Map

func TearUpWithEmptySD(svc enums.Svc, svcConf deploy.SvcConfImpl) {
	var o = new(sync.Once)
	v, ok := oncemapEmptySD.Load(svc.Name())
	if !ok {
		oncemapEmptySD.Store(svc.Name(), o)
	} else {
		o = v.(*sync.Once)
	}

	o.Do(func() {
		_ = os.Setenv(consts.EnvVarLogLevel, "debug")
		graceful.SetupSignal()

		if !isProjectRootDir() {
			wd, _ := os.Getwd()
			parentDir := filepath.Dir(filepath.Dir(wd))
			_ = os.Chdir(parentDir)
		}
		deploy.InitCfg(svc, svcConf)
		pkg.Setup(
			xlog.Init,
		)
		svccli.SetDefaultSD(abstract.Empty{})
		infra.Setup(
			//sd.Init(true),
			svccli.Init(true),
		)
	})
}

func isProjectRootDir() bool {
	_, err := os.Stat("go.mod")
	return err == nil
}

// TestCallCtx 它的traceId在多次使用时是同一个，若需要不同的traceId请直接调用 NewTestCallCtx()
var TestCallCtx = NewTestCallCtx(true, 1)
var TestCallCtxNoAuth = NewTestCallCtx(false, 0)

func NewTestCallCtx(isAuth bool, uid int64) context.Context {
	// 届时在server侧会在ctx中填冲一个uid=1的假用户
	md := metadata.Pairs(
		xgrpc.MdKeyTestCall, xgrpc.MdKeyFlagExist,
		xgrpc.MdKeyTraceId, util.NewKsuid(),
		xgrpc.MdKeyBizRemoteAddr, "127.0.0.1",
	)
	if isAuth && uid > 0 {
		md.Append(xgrpc.MdKeyFakeAuth, cast.ToString(uid))
	}
	return metadata.NewOutgoingContext(context.TODO(), md)
}

func NewTestCallCtxWithTimeout(duration time.Duration, uid int64) (context.Context, func()) {
	// 届时在server侧会在ctx中填冲一个uid=1的假用户
	md := metadata.Pairs(
		xgrpc.MdKeyTraceId, util.NewKsuid(),
		xgrpc.MdKeyBizRemoteAddr, "127.0.0.1",
	)
	if uid > 0 {
		md.Append(xgrpc.MdKeyFakeAuth, cast.ToString(uid))
	}
	ctx, cancel := context.WithTimeout(context.TODO(), duration)
	return metadata.NewOutgoingContext(ctx, md), cancel
}

var TestBaseExtReq = &commonpb.BaseExtReq{
	AppName:    "test_app",
	AppVersion: "1.0.0",
	Platform:   commonpb.SignInPlatform_SIP_APP,
	System:     commonpb.SignInSystem_SIS_Android,
	Language:   commonpb.Lang_CL_EN,
	Extension:  nil,
}

var TestBaseAdminReq = &adminpb.AdminBaseReq{
	UserAgent: "x",
	Platform:  commonpb.SignInPlatform_SIP_APP,
	System:    commonpb.SignInSystem_SIS_Android,
	Language:  commonpb.Lang_CL_EN,
	Extension: nil,
}

func GRPCHealthCheck(t *testing.T, commonpb enums.Svc, svcConf deploy.SvcConfImpl) {
	TearUp(commonpb, svcConf)
	defer TearDown()

	healthCli := svccli.NewCli(commonpb, func(conn *grpc.ClientConn) interface{} { return grpc_health_v1.NewHealthClient(conn) })
	cli := healthCli.Getter().(grpc_health_v1.HealthClient)

	response, err := cli.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{
		Service: commonpb.Name(),
	})
	if err != nil {
		panic(err)
	}

	assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, response.Status)
}
