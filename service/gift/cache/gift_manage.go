package cache

import (
	"context"
	"microsvc/model/svc/gift"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/giftpb"
	"microsvc/service/gift/dao"
	"microsvc/util/db"
	"microsvc/util/ujson"
	"time"

	"github.com/spf13/cast"

	"github.com/samber/lo"
)

type giftMgCtrlT struct {
	giftListExpire time.Duration
}

var GiftMgCtrl = giftMgCtrlT{
	giftListExpire: time.Minute * 15,
}

func (giftMgCtrlT) giftListKey() string {
	return GiftMapAllCacheKey
}

func (g giftMgCtrlT) List(ctx context.Context, scene giftpb.GiftScene) (list []*gift.GiftConf, err error) {
	return g.__listWithCond(ctx, &GiftParams{
		Scene: scene,
		Type:  nil,
		State: []giftpb.GiftState{giftpb.GiftState_GS_On},
	})
}

func (g giftMgCtrlT) ListAll(ctx context.Context) (list []*gift.GiftConf, err error) {
	return g.__listWithCond(ctx, &GiftParams{
		Scene: giftpb.GiftScene_GS_Unknown, // 表示all
		Type:  nil,
		State: nil,
	})
}

func (g giftMgCtrlT) GetOne(ctx context.Context, giftID int64, params *GiftParams) (obj *gift.GiftConf, err error) {
	key := g.giftListKey()
	buf, err := gift.R.HGet(ctx, key, cast.ToString(giftID)).Bytes()
	if db.IgnoreNilErr(err) != nil {
		return nil, xerr.WrapRedis(err)
	}
	defer func() {
		if obj == nil {
			return
		}
		list := g.filterGiftList([]*gift.GiftConf{obj}, params)
		if len(list) == 0 {
			obj = nil
		}
	}()
	if buf != nil {
		err = ujson.Unmarshal(buf, &obj)
		return
	}

	// load cache
	_, gmap, err := g.__loadList2Cache(ctx, false)
	if err != nil {
		return nil, err
	}
	return gmap[giftID], nil
}

func (g giftMgCtrlT) GetAvailableOne(ctx context.Context, giftID int64, scene giftpb.GiftScene) (obj *gift.GiftConf, err error) {
	obj, err = g.GetOne(ctx, giftID, &GiftParams{
		Scene: scene,
		Type:  nil,
		State: []giftpb.GiftState{giftpb.GiftState_GS_On},
	})
	if err != nil {
		return
	}
	if obj == nil {
		return nil, xerr.ErrGiftNotFound
	}
	return
}

type GiftParams struct {
	Scene giftpb.GiftScene
	Type  []giftpb.GiftType  // 空表示所有类型
	State []giftpb.GiftState // 空表示所有状态
}

func (g giftMgCtrlT) __listWithCond(ctx context.Context, params *GiftParams) (list []*gift.GiftConf, err error) {
	key := g.giftListKey()
	vmap, err := gift.R.HGetAll(ctx, key).Result()
	if db.IgnoreNilErr(err) != nil {
		return nil, xerr.WrapRedis(err)
	}
	defer func() {
		list = g.filterGiftList(list, params)
	}()

	if len(vmap) > 0 { // vmap 永远非nil
		for _, buf := range vmap {
			var obj *gift.GiftConf
			err = ujson.Unmarshal([]byte(buf), &obj)
			if err != nil {
				return nil, xerr.WrapRedis(err)
			}
			list = append(list, obj)
		}
		return
	}

	// load cache
	list, _, err = g.__loadList2Cache(ctx, false)
	return
}

func (g giftMgCtrlT) filterGiftList(list []*gift.GiftConf, params *GiftParams) (list2 []*gift.GiftConf) {
	for _, v := range list {
		if len(params.State) > 0 && !lo.Contains(params.State, v.State) {
			continue
		}
		if len(params.Type) > 0 && !lo.Contains(params.Type, v.Type) {
			continue
		}
		if params.Scene != giftpb.GiftScene_GS_Unknown && !lo.Contains(v.SupportedScenes, params.Scene) {
			continue
		}
		list2 = append(list2, v)
	}
	// 不需要排序，mysql查询时已经按照 price,id 排序
	return
}

func (g giftMgCtrlT) __loadList2Cache(ctx context.Context, refresh bool) (list []*gift.GiftConf, gmap map[int64]*gift.GiftConf, err error) {
	key := g.giftListKey()

	if refresh {
		err = gift.R.Del(ctx, key).Err()
		if err != nil {
			return
		}
	}
	list, err = dao.GiftConfDao.ListAllGiftConf(ctx)
	if err != nil {
		return
	}
	gmap = make(map[int64]*gift.GiftConf)
	if len(list) == 0 {
		return
	}
	var values []interface{}
	lo.ForEach(list, func(v *gift.GiftConf, _ int) {
		values = append(values, v.Id, ujson.MustMarshal(v))
		gmap[v.Id] = v
	})
	err = gift.R.HMSet(ctx, key, values...).Err()
	err = xerr.WrapRedis(err)
	return
}
