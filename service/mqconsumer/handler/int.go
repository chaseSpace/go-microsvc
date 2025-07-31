package handler

import (
	"context"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/mqconsumerpb"

	"go.uber.org/zap"
)

var IntCtrl mqconsumerpb.MqConsumerIntServer = new(intCtrl)

type intCtrl struct{}

func (t intCtrl) ReportMsg(ctx context.Context, req *mqconsumerpb.ReportMsgReq) (*commonpb.EmptyRes, error) {
	xlog.Info("report msg", zap.Any("req", req))
	return &commonpb.EmptyRes{}, nil
}
