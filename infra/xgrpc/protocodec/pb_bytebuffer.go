package protocodec

import (
	"fmt"

	"github.com/valyala/bytebufferpool"

	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"
)

const PBByteBuffer = "pb-bytebuffer"

func init() {
	encoding.RegisterCodec(codecPBByteBuffer{})
}

type codecPBByteBuffer struct{}

func (c codecPBByteBuffer) Marshal(v interface{}) ([]byte, error) {
	vv, ok := v.(proto.Message)
	if !ok {
		vb, ok := v.([]byte) // send by from gateway OR server response to gateway
		if !ok {
			return nil, fmt.Errorf(c.Name()+": failed to marshal, message is %T", v)
		}
		return vb, nil
	}
	// Send by sub services
	return proto.Marshal(vv)
}

func (c codecPBByteBuffer) Unmarshal(data []byte, v interface{}) error {
	vv, ok := v.(proto.Message)
	if !ok {
		vb, ok := v.(*bytebufferpool.ByteBuffer) // from gateway
		if !ok {
			return fmt.Errorf(c.Name()+": failed to unmarshal, message is %T", v)
		}
		_, err := vb.Write(data)
		return err
	}
	// Received on gateway
	return proto.Unmarshal(data, vv)
}

func (codecPBByteBuffer) Name() string {
	return PBByteBuffer
}
