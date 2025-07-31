package deploy

import (
	"microsvc/consts"
	"net/url"
	"strings"

	"github.com/samber/lo"
)

type DefaultAssets struct {
	BarCover   string `mapstructure:"bar_cover"`
	WineCover  string `mapstructure:"wine_cover"`
	EventCover string `mapstructure:"event_cover"`
}

func (a *XConfig) FullImgUrl(imgPath string) string {
	if imgPath == "" || strings.HasPrefix(imgPath, "http:") || strings.HasPrefix(imgPath, "https:") {
		return imgPath
	}
	u, _ := url.JoinPath(a.getImgHost(imgPath), imgPath)
	return u
}

func (a *XConfig) FullImgUrls(list []string) []string {
	return lo.Map(list, func(item string, index int) string {
		return a.FullImgUrl(item)
	})
}

func (a *XConfig) getImgHost(imgPath string) string {
	imgPath = strings.TrimPrefix(imgPath, "/")
	if strings.HasPrefix(imgPath, consts.AdminFileUploadReqPath) {
		return a.AdminSystem.ImgBaseUrl
	} else if strings.HasPrefix(imgPath, consts.MicroSvcFileUploadReqPath) {
		return a.ApiGateway.BaseUrl
	} else if strings.HasPrefix(imgPath, consts.ManualUploadReqPath) {
		return a.OfficialSite.BaseUrl
	}
	return ""
}
