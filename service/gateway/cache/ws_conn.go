package cache

import (
	"context"
	"microsvc/bizcomm/commgw"
	"microsvc/model/svc/gateway"
	"microsvc/pkg/xerr"
	"microsvc/util/ujson"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cast"
)

type wsCache struct {
}

var WsCache = wsCache{}

func (w wsCache) HashValueConn(connID, gwID string, detail commgw.LoginDetail) *commgw.HashValueConn {
	return &commgw.HashValueConn{
		ConnId:    connID,
		GatewayID: gwID,
		CreatedAt: time.Now().Unix(),
		Detail:    detail,
	}
}

func (w wsCache) SaveUserConnMap(ctx context.Context, uid int64, connSave *commgw.HashValueConn) error {
	err := gateway.R.HSet(ctx, CKeyHashUID2ConnID, cast.ToString(uid), ujson.MustMarshal(connSave)).Err()
	return xerr.WrapRedis(err)
}

func (w wsCache) GetConnByUID(ctx context.Context, uid int64) (conn *commgw.HashValueConn, err error) {
	return commgw.GetConnByUID(ctx, uid)
}

func (w wsCache) DeleteConnByUID(ctx context.Context, uid ...int64) error {
	keys := lo.Map(uid, func(item int64, _ int) string {
		return cast.ToString(item)
	})
	return gateway.R.HDel(ctx, CKeyHashUID2ConnID, keys...).Err()
}
