package logic_oss

import (
	"context"
	"fmt"
	"microsvc/consts"
	deploy2 "microsvc/deploy"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/service/thirdparty/deploy"
	"microsvc/service/thirdparty/logic_oss/thirdparty"
	"microsvc/service/thirdparty/logic_oss/thirdparty/tencent"
	"microsvc/util"
	"microsvc/util/uencode"
	"microsvc/util/ufile"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cast"
)

var provider thirdparty.OssAPI

func MustInit(cc *deploy.SvcConfig) {
	provider = &tencent.OssImpl{}
	provider.MustInit(cc.Oss.Tencent)
}

func uploadResource(ctx context.Context, uid int64, typ commonpb.OSSUploadType, buf []byte) (path, url string, err error) {
	dtStr := time.Now().Format(consts.DateToHMSLayout2)
	if uid > 0 {
		path, url, err = provider.UploadUserResource(ctx, strings.Join([]string{cast.ToString(uid), typ.String(), dtStr}, "/"), buf)
	} else {
		path, url, err = provider.UploadPublicResource(ctx, strings.Join([]string{typ.String(), dtStr}, "/"), buf)
	}
	return path, url, err
}

func checkUploadParams(ctx context.Context, uid int64, typ commonpb.OSSUploadType, buf []byte) (err error) {
	switch typ {
	case commonpb.OSSUploadType_OUT_Avatar,
		commonpb.OSSUploadType_OUT_Background,
		commonpb.OSSUploadType_OUT_Album,
		commonpb.OSSUploadType_OUT_AlbumCover:
		if uid < 1 {
			return xerr.ErrParams.New("Need provide uid for this type")
		}
		// 检查文件内容是否图片
		isImage, ext, err := ufile.IsImageFile(buf, ufile.CommonImageTypes...)
		if err != nil {
			return xerr.ErrInvalidFileType.AppendMsg(err.Error())
		}
		if !isImage {
			return xerr.ErrInvalidFileType.AppendMsg("Unsupported image type: %s", ext)
		}
	default:
		return xerr.ErrParams.New("Unsupported upload type")
	}
	return
}

type FileBizTypeConfT struct {
	MaxBytes ufile.FileSize // 1024 = 1KB
}

var FileBizTypeConf = map[thirdpartypb.FileBizType]*FileBizTypeConfT{
	thirdpartypb.FileBizType_FBT_Avatar: {
		MaxBytes: ufile.FSizeMB * 2,
	},
	thirdpartypb.FileBizType_FBT_BAR_BG: {
		MaxBytes: ufile.FSizeMB * 5,
	},
}

// 检查文件业务类型是否匹配buf信息（类型、大小）
func checkFileBizTypeMatchBufInfo(ctx context.Context, uid int64, typ thirdpartypb.FileBizType, bufBase64 string) (buf []byte, name string, err error) {
	buf, err = uencode.Base64Decode(bufBase64)
	if err != nil {
		return nil, "", xerr.ErrParams.New("Failed to decode base64 string of file buffer").Append(err)
	}
	// 检查文件大小
	s := FileBizTypeConf[typ]
	if s != nil && len(buf) > s.MaxBytes.Int() {
		return nil, "", xerr.ErrParams.New("File size too large, max: %s", s)
	}

	var requireImg bool
	switch typ {
	case thirdpartypb.FileBizType_FBT_Avatar,
		thirdpartypb.FileBizType_FBT_BAR_BG,
		thirdpartypb.FileBizType_FBT_BAR_Cover:
		requireImg = true
	}

	if requireImg { // bin检查
		isImage, ext, err := ufile.IsImageFile(buf, ufile.CommonImageTypes...)
		if err != nil {
			return nil, "", xerr.ErrInvalidFileType.Append(err)
		}
		if !isImage {
			return nil, "", xerr.ErrInvalidFileType.New("Unsupported image type: %s", ext)
		}
		name = genLocalUploadFileName(fmt.Sprintf(`%d@%s.`, uid, util.NewKsuid()) + ext)
	}
	// 其他文件 也要生成对应的名字
	return
}

func genLocalUploadFileName(srcName string) string {
	return fmt.Sprintf(`%s@%s`, time.Now().Format(consts.DateToHMSLayout3), srcName)
}

func localUpload(ctx context.Context, uid int64, BizType thirdpartypb.FileBizType, base64 string) (path, accessUri string, err error) {
	buf, name, err := checkFileBizTypeMatchBufInfo(ctx, uid, BizType, base64)
	if err != nil {
		return "", "", err
	}
	uploads := deploy.ThirdpartyConf.Oss.LocalUploadDir
	date := time.Now().Format("20060102")

	pathArr := []string{uploads, BizType.String(), date, name}
	// e.g. "./micro_uploads/FBT_Avatar/20250519/165710@1@2xJ8iWAAp7jmXapBlZYSx7Djk5d.png"
	writePath := strings.Join(pathArr, "/")
	_url, _ := url.JoinPath(deploy2.XConf.ApiGateway.BaseUrl, writePath)

	err = os.MkdirAll(filepath.Dir(writePath), os.ModePerm)
	if err != nil {
		return "", "", xerr.ErrWriteFileFailed.Append(err)
	}
	err = os.WriteFile(writePath, buf, os.ModePerm)
	if err != nil {
		return "", "", xerr.ErrWriteFileFailed.Append(err)
	}
	return writePath, _url, nil
}
