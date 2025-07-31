package protoutil

import (
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/util/ujson"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// WrapExtResponse
/*
old external grpc response(err==nil):

	{
	  "a": 1
	}

new external grpc response:

	{
	   "code": 200,
	   "msg": "OK",
	   "data": {"a": 1},
	}
*/

// HTTPResp 这个结构体是为了取代 commonpb.HTTPResp ,后者在JSON序列化时存在一些无法解决的问题！
type HTTPResp struct {
	Code           int32    `json:"code"`
	Msg            string   `json:"msg"`
	FromGateway    bool     `json:"from_gateway,omitempty"`
	Data           any      `json:"data"`
	PassedServices []string `json:"passed_services,omitempty"` // error 时用于排查
}

type WrapExtRes struct{}

func (WrapExtRes) wrapToPB(data interface{}, err error) (proto.Message, error) {
	var v *anypb.Any
	var err2 error
	if data != nil {
		d2 := data.(proto.Message)
		v, err2 = anypb.New(d2)
		if err2 != nil {
			return nil, errors.Wrap(err2, "WrapExtResponse")
		}
	}
	res := &commonpb.HTTPResp{
		Code:        xerr.ErrNil.Code,
		Msg:         xerr.ErrNil.Msg,
		FromGateway: false,
		Data:        v,
	}
	if err != nil {
		xe := xerr.ToXErr(err)
		res.Code = xe.Code
		res.Msg = xe.Msg
		res.PassedServices = xe.PassedServices
	}
	return res, nil
}

func (WrapExtRes) wrapToJSON(data interface{}, err error, fromGw bool) ([]byte, error) {
	res := &HTTPResp{
		Code:        xerr.ErrNil.Code,
		Msg:         xerr.ErrNil.Msg,
		FromGateway: fromGw,
		Data:        data,
	}
	if err != nil {
		xe := xerr.ToXErr(err)
		res.Code = xe.Code
		res.Msg = xe.FormatMsg()
		res.PassedServices = xe.PassedServices
	}
	return ujson.Marshal(res)
}

func (w WrapExtRes) OnService(data interface{}, err error, isPB bool) (interface{}, error) {
	if isPB {
		return w.wrapToPB(data, err)
	}
	return w.wrapToJSON(data, err, false)
}

func (w WrapExtRes) OnGateway(data interface{}, err error, fromGw, isPB bool) ([]byte, error) {
	if isPB {
		pb, err := w.wrapToPB(data, err)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(pb)
	}
	return w.wrapToJSON(data, err, fromGw)
}

func WrapResponseOnService(data interface{}, err error, isPB bool) (interface{}, error) {
	return WrapExtRes{}.OnService(data, err, isPB)
}
func WrapResponseOnGateway(data interface{}, err error, fromGw, isPB bool) ([]byte, error) {
	return WrapExtRes{}.OnGateway(data, err, fromGw, isPB)
}
