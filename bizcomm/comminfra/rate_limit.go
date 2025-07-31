package comminfra

import (
	"context"
	"fmt"
	"microsvc/bizcomm/commcache"
	"microsvc/enums"
	"microsvc/model/svc/micro_svc"
	"microsvc/pkg/xerr"
	"microsvc/pkg/xlog"
	"microsvc/xvendor/xratelimit"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	DefaultSingleAPIMaxQPSByIP  = 10
	DefaultSingleAPIMaxQPSByUID = 5
)

var (
	defaultAPIRateLimitConf = &micro_svc.APIRateLimitConf{
		//Svc:           "",
		//APIPath:       "",
		MaxQPSByIP:  DefaultSingleAPIMaxQPSByIP,
		MaxQPSByUID: DefaultSingleAPIMaxQPSByUID,
		State:       micro_svc.APIRateLimitConfStateEnabled,
	}
)

type APIRateLimitConfProvider struct {
	mu      sync.Mutex
	confMap *sync.Map
	runOnce sync.Once

	refreshInterval time.Duration
	limiter         *xratelimit.RateLimiter
}

func NewAPIRateLimitConfProvider(isOpen bool, limiter *xratelimit.RateLimiter) *APIRateLimitConfProvider {
	if !isOpen { // 关闭限速
		defaultAPIRateLimitConf.MaxQPSByIP = 0
		defaultAPIRateLimitConf.MaxQPSByUID = 0
	}
	return &APIRateLimitConfProvider{
		confMap:         &sync.Map{},
		refreshInterval: time.Minute * 5,
		limiter:         limiter,
	}
}

func (a *APIRateLimitConfProvider) CheckByIP(handlerCtx context.Context, apiPath, IP string, handler func() (interface{}, error)) (interface{}, error) {
	cc := a.getOne(apiPath)
	if cc.MaxQPSByIP < 1 {
		return handler()
	}

	key := fmt.Sprintf(commcache.CacheKeyRateLimitByIP, IP, apiPath)
	allow, err := a.limiter.Allow(handlerCtx, key, cc.MaxQPSByIP)
	if err != nil {
		return nil, err
	}
	if !allow {
		return nil, xerr.ErrTooManyRequests
	}
	return handler()
}

func (a *APIRateLimitConfProvider) CheckByUID(handlerCtx context.Context, apiPath string, uid int64, handler func() (interface{}, error)) (interface{}, error) {
	cc := a.getOne(apiPath)
	if cc.MaxQPSByUID < 1 {
		return handler()
	}

	key := fmt.Sprintf(commcache.CacheKeyRateLimitByUID, uid, apiPath)
	allow, err := a.limiter.Allow(handlerCtx, key, cc.MaxQPSByUID)
	if err != nil {
		return nil, err
	}
	if !allow {
		return nil, xerr.ErrTooManyRequests
	}
	return handler()
}

func (a *APIRateLimitConfProvider) getOne(apiPath string) *micro_svc.APIRateLimitConf {
	cc, _ := a.confMap.Load(apiPath)
	if cc == nil {
		return defaultAPIRateLimitConf
	}
	return cc.(*micro_svc.APIRateLimitConf)
}

func (a *APIRateLimitConfProvider) AutoRefresh() {
	a.runOnce.Do(func() {
		a.refresh()
		go a.runTimer()
	})
}

func (a *APIRateLimitConfProvider) refresh() {
	list, err := micro_svc.GetOpenedAPIRateLimitConf()
	if err != nil {
		xlog.Error("APIRateLimitConfProvider refresh failed", zap.Error(err))
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	length := 0

	a.confMap.Range(func(key, value interface{}) bool {
		length++
		return true
	})
	if length == 0 && len(list) == 0 {
		return
	}

	// 申请新的map来替换
	newMap := sync.Map{}
	for _, row := range list {
		if row.Svc == "" || row.APIPath == "" {
			xlog.Error("APIRateLimitConfProvider read invalid conf", zap.Any("row", row))
			continue
		}
		newMap.Store(genGRPCFullMethod(row.Svc, row.APIPath), row)
	}
	a.confMap = &newMap
}

func (a *APIRateLimitConfProvider) runTimer() {
	timer := time.NewTicker(a.refreshInterval)
	for {
		<-timer.C
		xlog.Info("APIRateLimitConfProvider refresh")
		a.refresh()
	}
}

// genGRPCFullMethod
// - svc: user
// - fullMethod: UserExt/GetUserInfo
func genGRPCFullMethod(svc enums.Svc, fullMethod string) string {
	f := fmt.Sprintf("/svc.%s.%s", svc, fullMethod)
	return f
}
