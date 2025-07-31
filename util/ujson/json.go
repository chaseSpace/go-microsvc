package ujson

import (
	"io"
	"microsvc/pkg/xerr"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	jsoniter "github.com/json-iterator/go"
)

func Marshal(src interface{}) ([]byte, error) {
	buf, err := jsoniter.Marshal(src)
	if err != nil {
		return nil, xerr.ErrJSONMarshal.Append(err)
	}
	return buf, nil
}

func MustMarshal(src interface{}) []byte {
	buf, err := jsoniter.Marshal(src)
	if err != nil {
		panic(xerr.ErrJSONMarshal.Append(err))
	}
	return buf
}

func MustMarshal2Str(src interface{}) string {
	return string(MustMarshal(src))
}

func Unmarshal(buf []byte, dst interface{}) error {
	err := jsoniter.Unmarshal(buf, dst)
	if err != nil {
		return xerr.ErrJSONUnmarshal.Append(err)
	}
	return nil
}

func MustUnmarshal(buf []byte, dst interface{}) {
	err := Unmarshal(buf, dst)
	if err != nil {
		panic(xerr.ErrJSONUnmarshal.Append(err))
	}
}

func UnmarshalReader(reader io.Reader, dst interface{}) error {
	err := jsoniter.NewDecoder(reader).Decode(dst)
	if err != nil {
		return xerr.ErrJSONUnmarshal.Append(err)
	}
	return nil
}

// Deprecated: 这个方法会将proto类型int64的字段序列化为string，许多场景不可接受！
func ProtoJsonMarshal(src proto.Message, emitDefaultVal bool) ([]byte, error) {
	return protojson.MarshalOptions{
		UseEnumNumbers:    true,
		UseProtoNames:     true,
		EmitDefaultValues: emitDefaultVal,
	}.Marshal(src)
}

// Deprecated: 这个方法会将proto类型int64的字段序列化为string，许多场景不可接受！
func MustProtoJsonMarshal(src proto.Message, emitDefaultVal bool) []byte {
	buf, err := ProtoJsonMarshal(src, emitDefaultVal)
	if err != nil {
		panic(xerr.ErrJSONMarshal.Append(err))
	}
	return buf
}
