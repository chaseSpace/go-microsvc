package ufile

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers"
	"github.com/h2non/filetype/types"
	"github.com/samber/lo"
)

var (
	CommonImageTypes = []types.Type{matchers.TypeJpeg, matchers.TypePng, matchers.TypeWebp, matchers.TypeJpeg2000}
)

type FileSize float64

func (f FileSize) String() string {
	if f < 1024 {
		return fmt.Sprintf("%fB", f)
	}
	if f < 1024*1024 {
		return fmt.Sprintf("%.1fKB", f/1024)
	}
	return fmt.Sprintf("%.1fMB", f/1024/1024)
}

func (f FileSize) Int() int {
	return int(f)
}

const (
	FSizeKB = FileSize(1024)
	FSizeMB = FileSize(1024 * 1024)
	FSizeGB = FileSize(1024 * 1024 * 1024)
)

func IsImageFile(buf []byte, typ ...types.Type) (bool, string, error) {
	if len(typ) == 0 {
		typ = CommonImageTypes
	}
	kind, err := filetype.Match(buf)
	if err != nil {
		return false, "", err
	}
	return lo.Contains(typ, kind), kind.Extension, nil
}

func AssertPathsExist(paths ...string) {
	for _, path := range paths {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			panic(fmt.Sprintf("file does not exist: " + path))
		}

	}
}

func MustRead(file string) (buf []byte) {
	buf, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return
}

func MustReadToBase64(file string) (bufEncoded string) {
	buf := MustRead(file)
	return base64.StdEncoding.EncodeToString(buf)
}
