package user

import (
	"microsvc/consts"
	"microsvc/enums"
	"microsvc/infra/svccli/rpc"
	"microsvc/protocol/svc/adminpb"
	"microsvc/protocol/svc/commonpb"
	"microsvc/service/admin/deploy"
	"microsvc/test/tbase"
	"testing"
)

func init() {
	tbase.TearUp(enums.SvcAdmin, deploy.AdminConf)
}

func TestConfigCenterGet(t *testing.T) {
	defer tbase.TearDown()

	rpc.Admin().ConfigCenterGet(tbase.TestCallCtx, &adminpb.ConfigCenterGetReq{
		Keys: []string{consts.AdminCCKeySpiderInsCookie},
	})
}

func TestConfigCenterUpdateInt(t *testing.T) {
	defer tbase.TearDown()

	rpc.Admin().ConfigCenterUpdateInt(tbase.TestCallCtx, &adminpb.ConfigCenterUpdateIntReq{
		Item: &commonpb.ConfigItemCore{
			Key:                "2",
			Name:               "1",
			Value:              "1",
			AllowProgramUpdate: true,
		},
	})
}

func TestSwitchCenterGet(t *testing.T) {
	defer tbase.TearDown()

	rpc.Admin().SwitchCenterGet(tbase.TestCallCtx, &adminpb.SwitchCenterGetReq{
		Keys: []string{"123"},
	})
}
