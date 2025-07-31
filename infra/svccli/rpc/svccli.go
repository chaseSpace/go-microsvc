//go:build !k8s

package rpc

import (
	"microsvc/enums"
	"microsvc/infra/svccli"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/currencypb"
	"microsvc/protocol/svc/giftpb"
	"microsvc/protocol/svc/momentpb"
	"microsvc/protocol/svc/mqconsumerpb"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/protocol/svc/userpb"

	"google.golang.org/grpc"
)

// If you use this file, Service client use ServiceDiscovery method
// to get service target address, ServiceDiscovery could be implemented
// by Consul/etcd/ZooKeeper/Nacos etc.

var (
	adminCli      = svccli.NewCli(enums.SvcAdmin, func(conn *grpc.ClientConn) interface{} { return adminpb.NewAdminIntClient(conn) })
	userCli       = svccli.NewCli(enums.SvcUser, func(conn *grpc.ClientConn) interface{} { return userpb.NewUserIntClient(conn) })
	thirdpartyCli = svccli.NewCli(enums.SvcThirdparty, func(conn *grpc.ClientConn) interface{} { return thirdpartypb.NewThirdpartyIntClient(conn) })
	currencyCli   = svccli.NewCli(enums.SvcCurrency, func(conn *grpc.ClientConn) interface{} { return currencypb.NewCurrencyIntClient(conn) })
	giftCli       = svccli.NewCli(enums.SvcGift, func(conn *grpc.ClientConn) interface{} { return giftpb.NewGiftIntClient(conn) })
	momentCli     = svccli.NewCli(enums.SvcMoment, func(conn *grpc.ClientConn) interface{} { return momentpb.NewMomentIntClient(conn) })
	mqconsumer    = svccli.NewCli(enums.SvcMqConsumer, func(conn *grpc.ClientConn) interface{} { return mqconsumerpb.NewMqConsumerIntClient(conn) })
)

func Admin() adminpb.AdminIntClient {
	return adminCli.Getter().(adminpb.AdminIntClient)
}

func User() userpb.UserIntClient {
	return userCli.Getter().(userpb.UserIntClient)
}

func Thirdparty() thirdpartypb.ThirdpartyIntClient {
	return thirdpartyCli.Getter().(thirdpartypb.ThirdpartyIntClient)
}

func Currency() currencypb.CurrencyIntClient {
	return currencyCli.Getter().(currencypb.CurrencyIntClient)
}

func Gift() giftpb.GiftIntClient {
	return giftCli.Getter().(giftpb.GiftIntClient)
}

func Moment() momentpb.MomentIntClient {
	return momentCli.Getter().(momentpb.MomentIntClient)
}

func MqConsumer() mqconsumerpb.MqConsumerIntClient {
	return mqconsumer.Getter().(mqconsumerpb.MqConsumerIntClient)
}
