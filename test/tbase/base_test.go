package tbase

import (
	"context"
	"microsvc/enums"
	"microsvc/infra/svccli/rpcext"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/userpb"
	deploy2 "microsvc/service/user/deploy"
	"testing"

	"github.com/hashicorp/go-uuid"
	"github.com/segmentio/ksuid"
)

func TestNoRPCClient(t *testing.T) {
	TearUpWithEmptySD(enums.SvcUser, deploy2.UserConf)
	defer TearDown()

	_, err := rpcext.User().GetUserInfo(context.TODO(), &userpb.GetUserInfoReq{
		Base: nil,
		Uids: nil,
	})
	if !xerr.ErrAPIUnavailable.Is(err) {
		t.Errorf("case 1: err is not ErrAPIUnavailable: %v", err)
	}
}

// Run userpb svc first.
func TestHaveRPCClient(t *testing.T) {
	TearUp(enums.SvcUser, deploy2.UserConf)
	defer TearDown()

	rsp, err := rpcext.User().GetUserInfo(context.TODO(), &userpb.GetUserInfoReq{
		Base: nil,
		Uids: []int64{1},
	})
	if err != nil {
		t.Errorf("case 1: err: %v", err)
	} else {
		if len(rsp.Umap) != 1 {
			t.Errorf("case 1: err rsp: %+v", rsp.Umap)
		}
	}
}

func TestUUID(t *testing.T) {
	imap := make(map[string]interface{})
	for i := 0; i < 10000000; i++ {
		s, err := uuid.GenerateUUID()
		if err != nil {
			t.Fatalf(err.Error())
		}
		if imap[s] != nil {
			t.Fatal("exists", i)
		}
		imap[s] = 1
		println(i, s)
	}
}

func TestKsuid(t *testing.T) {
	imap := make(map[string]interface{})
	// https://github.com/segmentio/ksuid
	// 生成  一种可按生成时间排序、固定20 bytes的 唯一id；无碰撞、无协调、无依赖
	// - 按时间戳按字典顺序排序
	// - base62 编码的文本表示，url友好，复制友好

	s := ksuid.New()
	for i := 0; i < 5000000; i++ {
		s = s.Next()
		if imap[s.String()] != nil {
			t.Fatal("exists", i)
		}
		imap[s.String()] = 1
		println(i, s.String())
	}
}
