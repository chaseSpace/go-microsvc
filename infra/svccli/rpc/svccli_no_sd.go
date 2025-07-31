//go:build k8s

package rpc

import (
	"microsvc/enums"
	"microsvc/infra/sd"
	"microsvc/infra/xgrpc"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/currencypb"
	"microsvc/protocol/svc/giftpb"
	"microsvc/protocol/svc/momentpb"
	"microsvc/protocol/svc/mqconsumerpb"
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

func Admin() adminpb.AdminIntClient {
	return adminpb.NewAdminIntClient(getGRPCClient(enums.SvcAdmin))
}

func User() userpb.UserIntClient {
	return userpb.NewUserIntClient(getGRPCClient(enums.SvcUser))
}

func Thirdparty() thirdpartypb.ThirdpartyIntClient {
	return thirdpartypb.NewThirdpartyIntClient(getGRPCClient(enums.SvcThirdparty))
}

func Currency() currencypb.CurrencyIntClient {
	return currencypb.NewCurrencyIntClient(getGRPCClient(enums.SvcCurrency))
}

func Gift() giftpb.GiftIntClient {
	return giftpb.NewGiftIntClient(getGRPCClient(enums.SvcGift))
}

func Moment() momentpb.MomentIntClient {
	return momentpb.NewMomentIntClient(getGRPCClient(enums.SvcMoment))
}

func MqConsumer() mqconsumerpb.MqConsumerIntClient {
	return mqconsumerpb.NewMqConsumerIntClient(getGRPCClient(enums.SvcMqConsumer))
}
