package cache

import (
	"context"
	"fmt"
	"microsvc/model/svc/currency"
	"microsvc/pkg/xerr"
	"microsvc/service/currency/dao"
	"microsvc/util"
	"microsvc/util/db"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type goldCtrlT struct {
	goldAccountExpire time.Duration
}

var GoldCtrl = goldCtrlT{
	goldAccountExpire: time.Minute * 5,
}

func (goldCtrlT) goldAccountKey(uid int64) string {
	return fmt.Sprintf(goldAccountCacheKey, uid)
}

func (g goldCtrlT) GetGoldAccount(ctx context.Context, uid int64) (row *currency.GoldAccount, err error) {
	key := g.goldAccountKey(uid)

	var buf []byte
	buf, err = currency.R.Get(ctx, key).Bytes()
	if db.IgnoreNilErr(err) != nil {
		return nil, xerr.WrapRedis(err)
	}
	if buf != nil {
		_ = jsoniter.Unmarshal(buf, &row)
		return
	}
	row2, err := dao.GoldAccountDao.GetGoldAccount(ctx, nil, uid)
	if err != nil {
		return nil, err
	}
	// Set cache
	err = currency.R.Set(ctx, key, util.ToJson(&row2), g.goldAccountExpire).Err()
	return &row2, err
}

func (g goldCtrlT) ClearGoldAccount(ctx context.Context, uid ...int64) (err error) {
	for _, u := range uid {
		key := g.goldAccountKey(u)
		err = currency.R.Del(ctx, key).Err()
		if err != nil {
			return
		}
	}
	return
}
