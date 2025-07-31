package commrpc

import (
	"context"
	"microsvc/infra/svccli/rpc"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/userpb"

	"github.com/samber/lo"
)

type Object interface {
	GetUIDs() []int64
	SetUser(user ...*commonpb.User)
}

func PopulateUserBase[T Object](ctx context.Context, list []T) error {
	uids := make([]int64, 0)
	for _, item := range list {
		uids = append(uids, item.GetUIDs()...)
	}

	res, err := rpc.User().GetUserInfoInt(ctx, &userpb.GetUserInfoIntReq{
		Uids:             lo.Uniq(uids),
		PopulateNotfound: true,
	})
	if err != nil {
		return err
	}

	users := make([]*commonpb.User, 0)
	for _, v := range uids {
		users = append(users, res.Umap[v])
	}
	lo.ForEach(list, func(item T, index int) {
		item.SetUser(users...)
	})
	return nil
}
