package commadmin

import (
	"context"
	"microsvc/infra/svccli/rpc"
	"microsvc/protocol/svc/adminpb"
)

func SwitchCenterGetOne(ctx context.Context, skey SwitchKey) (item *SwitchItem, err error) {
	res, err := rpc.Admin().SwitchCenterGet(ctx, &adminpb.SwitchCenterGetReq{Keys: []string{string(skey)}})
	if err != nil {
		return nil, err
	}
	if v := res.Smap[string(skey)]; v != nil {
		return &SwitchItem{SwitchItem: v}, nil
	}
	// 当管理后台未配置时，返回默认值
	return DefaultSwitchValue(skey), nil
}
