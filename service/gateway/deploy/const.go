package deploy

import (
	"fmt"
	"microsvc/consts"
	"microsvc/enums"
)

const (
	PathProxyPrefix = "/forward/"
	PathPing        = "/ping"
	PathWS          = "/ws" // websocket
	PathUploads     = "/" + consts.MicroSvcFileUploadReqPath

	CtxKeyFromGateway = "from-gateway"
)

type ForwardDestination struct {
	Svc  enums.Svc
	Path string
}

func (v ForwardDestination) GetGRPCPath() string {
	return fmt.Sprintf("/svc.%s.%[1]sExt/%s", v.Svc.Name(), v.Path)
}
