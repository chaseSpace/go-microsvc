//go:build k8s

package rpcext

import (
	"microsvc/enums"
	"microsvc/infra/sd"
	"microsvc/infra/xgrpc"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/currencypb"
	"microsvc/protocol/svc/friendpb"
	"microsvc/protocol/svc/giftpb"
	"microsvc/protocol/svc/momentpb"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/protocol/svc/userpb"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// If you use this file, service client is directly use DNS name as service target address (e.g. in K8s environment).

func getGRPCClient(svc enums.Svc) *grpc.ClientConn {
	target := sd.GetSvcTargetInK8s(svc)
	conn, err := xgrpc.NewGRPCClient(target, svc.Name())
	if err != nil {
		xlog.Error("getGRPCClient", zap.Error(err))
		return xgrpc.NewInvalidGRPCConn(svc.Name())
	}
	return conn
}

func User() userpb.UserExtClient {
	return userpb.NewUserExtClient(getGRPCClient(enums.SvcUser))
}

func Admin() adminpb.AdminExtClient {
	return adminpb.NewAdminExtClient(getGRPCClient(enums.SvcAdmin))
}

func Thirdparty() thirdpartypb.ThirdpartyExtClient {
	return thirdpartypb.NewThirdpartyExtClient(getGRPCClient(enums.SvcThirdparty))
}

func Friend() friendpb.FriendExtClient {
	return friendpb.NewFriendExtClient(getGRPCClient(enums.SvcFriend))
}

func Currency() currencypb.CurrencyExtClient {
	return currencypb.NewCurrencyExtClient(getGRPCClient(enums.SvcCurrency))
}

func Gift() giftpb.GiftExtClient {
	return giftpb.NewGiftExtClient(getGRPCClient(enums.SvcGift))
}

func Moment() momentpb.MomentExtClient {
	return momentpb.NewMomentExtClient(getGRPCClient(enums.SvcMoment))
}
