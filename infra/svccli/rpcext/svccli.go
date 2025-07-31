//go:build !k8s

package rpcext

import (
	"microsvc/enums"
	"microsvc/infra/svccli"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/currencypb"
	"microsvc/protocol/svc/friendpb"
	"microsvc/protocol/svc/giftpb"
	"microsvc/protocol/svc/momentpb"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/protocol/svc/userpb"

	"google.golang.org/grpc"
)

// If you use this file, Service client use ServiceDiscovery method
// to get service target address, ServiceDiscovery could be implemented
// by Consul/etcd/ZooKeeper/Nacos etc.

var (
	userCli       = svccli.NewCli(enums.SvcUser, func(conn *grpc.ClientConn) interface{} { return userpb.NewUserExtClient(conn) })
	adminCli      = svccli.NewCli(enums.SvcAdmin, func(conn *grpc.ClientConn) interface{} { return adminpb.NewAdminExtClient(conn) })
	thirdpartyCli = svccli.NewCli(enums.SvcThirdparty, func(conn *grpc.ClientConn) interface{} { return thirdpartypb.NewThirdpartyExtClient(conn) })
	friendCli     = svccli.NewCli(enums.SvcFriend, func(conn *grpc.ClientConn) interface{} { return friendpb.NewFriendExtClient(conn) })
	currencyCli   = svccli.NewCli(enums.SvcCurrency, func(conn *grpc.ClientConn) interface{} { return currencypb.NewCurrencyExtClient(conn) })
	giftCli       = svccli.NewCli(enums.SvcGift, func(conn *grpc.ClientConn) interface{} { return giftpb.NewGiftExtClient(conn) })
	momentCli     = svccli.NewCli(enums.SvcMoment, func(conn *grpc.ClientConn) interface{} { return momentpb.NewMomentExtClient(conn) })
)

func User() userpb.UserExtClient {
	return userCli.Getter().(userpb.UserExtClient)
}

func Admin() adminpb.AdminExtClient {
	return adminCli.Getter().(adminpb.AdminExtClient)
}

func Thirdparty() thirdpartypb.ThirdpartyExtClient {
	return thirdpartyCli.Getter().(thirdpartypb.ThirdpartyExtClient)
}

func Friend() friendpb.FriendExtClient {
	return friendCli.Getter().(friendpb.FriendExtClient)
}

func Currency() currencypb.CurrencyExtClient {
	return currencyCli.Getter().(currencypb.CurrencyExtClient)
}

func Gift() giftpb.GiftExtClient {
	return giftCli.Getter().(giftpb.GiftExtClient)
}

func Moment() momentpb.MomentExtClient {
	return momentCli.Getter().(momentpb.MomentExtClient)
}
