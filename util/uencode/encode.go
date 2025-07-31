package uencode

import (
	"encoding/base64"

	"github.com/pkg/errors"
)

func Base64Encode(src []byte) (dst []byte, err error) {
	if len(src) == 0 {
		return nil, errors.New("nil input bytes")
	}
	dst = make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	base64.StdEncoding.Encode(dst, src)
	return
}

func Base64Decode(src string) (dst []byte, err error) {
	if len(src) == 0 {
		return nil, errors.New("nil input bytes")
	}
	dst, err = base64.StdEncoding.DecodeString(src)
	return dst, err
}
