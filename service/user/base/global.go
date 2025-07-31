package base

import (
	"context"
	"fmt"
	"microsvc/bizcomm/commuser"
	"microsvc/model/svc/user"
	"microsvc/service/user/cache"
	"microsvc/service/user/dao"
	"microsvc/service/user/deploy"
	"microsvc/util"
	"microsvc/util/ulock"
	"microsvc/xvendor/genuserid"

	"github.com/dlclark/regexp2"
	wxcache "github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/miniprogram"
	configmini "github.com/silenceper/wechat/v2/miniprogram/config"
	"github.com/silenceper/wechat/v2/officialaccount"
	"github.com/silenceper/wechat/v2/officialaccount/config"
)

// 一些子服务需要全局使用的资源在这里初始化
var (
	UidGenerator genuserid.UIDGenerator
	PhoneTool    = commuser.PhoneTool
	WxAppCli     *officialaccount.OfficialAccount // 微信原生使用公众号一致的API
	WxMiniCli    *miniprogram.MiniProgram
)

// MinUserId 新分配的靓号需要小于这个值，以完全避免与UID冲突
// - 一旦确定，不能改变
const MinUserId = 100000

func MustInit() {
	g := globalObjectCtrl{}
	util.AssertNilErr(g.InitUidGenerator())
	util.AssertNilErr(g.InitWxAppCli())
	util.AssertNilErr(g.InitWMiniCli())
}

type globalObjectCtrl struct {
}

func (g globalObjectCtrl) InitWxAppCli() error {
	cc := deploy.UserConf.WxApp
	cfg := config.Config{
		AppID:          cc.Appid,
		AppSecret:      cc.AppSecret,
		Token:          "",
		EncodingAESKey: "",
		Cache:          wxcache.NewRedis(context.Background(), g.newWxPkgRedisOpt()),
	}
	WxAppCli = officialaccount.NewOfficialAccount(&cfg)
	return nil
}

func (g globalObjectCtrl) InitWMiniCli() error {
	cc := deploy.UserConf.WxMini
	cfg := configmini.Config{
		AppID:          cc.Appid,
		AppSecret:      cc.AppSecret,
		Token:          "",
		EncodingAESKey: "",
		Cache:          wxcache.NewRedis(context.Background(), g.newWxPkgRedisOpt()),
	}
	WxMiniCli = miniprogram.NewMiniProgram(&cfg)
	return nil
}

func (globalObjectCtrl) newWxPkgRedisOpt() *wxcache.RedisOpts {
	return &wxcache.RedisOpts{
		Host:        user.R.Config().Addr,
		Password:    user.R.Config().Password,
		Database:    user.R.Config().DB,
		MaxIdle:     0,
		MaxActive:   0,
		IdleTimeout: 0,
	}
}

func (globalObjectCtrl) InitUidGenerator() error {
	// 预设需要跳过的靓号模式
	skipPattern := []string{
		`(\d)\1(\d)\2$`, // aabb结尾模式
		`(\d)\1{2}$`,    // aaa结尾模式，包含3个以上a结尾
		`(\d)\1{3}`,     // aaaa模式，包含4个以上a连续
	}
	skipFn := func(id uint64) (bool, error) {
		for _, p := range skipPattern {
			r := regexp2.MustCompile(p, 0) // 标准库regex不支持命名分组，所以第三方re库
			match, _ := r.MatchString(fmt.Sprintf("%d", id))
			if match {
				return true, nil
			}
		}
		return false, nil
	}

	locker := ulock.NewDLock("UidGenerator", user.R.Client)
	pool := cache.NewUidQueuedPool("UidGenerator", user.R.Client)

	getMaxUid := func() (uint64, error) {
		id, err := dao.GetMaxUid(context.Background())
		if err == nil && id < 1 {
			id = MinUserId
		}
		return id, err
	}

	var opts = []genuserid.Option{
		genuserid.WithSkipFunc(skipFn),
	}
	UidGenerator = genuserid.NewUidGenerator(locker, pool, getMaxUid, opts...)
	return nil
}
