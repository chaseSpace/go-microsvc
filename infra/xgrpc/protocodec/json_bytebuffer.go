package protocodec

import (
	"errors"
	"fmt"
	"microsvc/protocol/svc/commonpb"
	"microsvc/util/ujson"

	"github.com/valyala/bytebufferpool"
	"google.golang.org/protobuf/proto"

	"google.golang.org/grpc/encoding"
)

const JSONByteBuffer = "json-bytebuffer"

func init() {
	encoding.RegisterCodec(codecBytes{})
}

type codecBytes struct{}

func (c codecBytes) Marshal(v interface{}) ([]byte, error) {
	vv, ok := v.(*commonpb.HTTPResp)
	if !ok {
		vb, ok := v.([]byte) // send by from gateway OR server response to gateway
		if !ok {
			return nil, fmt.Errorf(c.Name()+": failed to marshal, message is %T", v)
		}
		return vb, nil
	}
	// Send by sub services
	return ujson.Marshal(vv)
}

// Unmarshal 如这个方法报错不会运行任何grpc拦截器
// runtime 上一跳
// _AdminExt_ListUser_Handler （具体API handler）
// reply, appErr := md.Handler(info.serviceImpl, ctx, df, s.opts.unaryInt)
func (c codecBytes) Unmarshal(data []byte, v interface{}) error {
	vv, ok := v.(proto.Message)
	if !ok {
		vb, ok := v.(*bytebufferpool.ByteBuffer) // from gateway
		if !ok {
			return fmt.Errorf(c.Name()+": failed to unmarshal, message is %T", v)
		}
		_, err := vb.Write(data)
		return err
	}
	// Received on sub services
	err := ujson.Unmarshal(data, vv)
	if err != nil {
		// Ignoring the origin err. Don't use `xerr` type, it would be captured by grpc Client interceptor `ExtractGRPCErr`
		return errors.New(fmt.Sprintf("invalid request body. Please refer to protobuf type `%T`", vv))
	}
	return nil
}

func (codecBytes) Name() string {
	return JSONByteBuffer
}
