package i18n

import (
	"microsvc/pkg/i18n/langimpl"
	"microsvc/protocol/svc/commonpb"
)

var defaultLangAPI = langimpl.English{}

var registerMap = map[commonpb.Lang]langimpl.LangAPI{}

func init() {
	ss := []langimpl.LangAPI{
		langimpl.CN, langimpl.EN,
	}
	for _, v := range ss {
		registerMap[v.Lang()] = v
	}
}

func TranslateError(lang commonpb.Lang, id langimpl.Code, subId langimpl.SubId) string {
	api := registerMap[lang]
	if api == nil {
		return defaultLangAPI.Translate(langimpl.SceneError, id, subId)
	}
	return api.Translate(langimpl.SceneError, id, subId)
}
