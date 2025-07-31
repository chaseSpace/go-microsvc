package xerr

import "microsvc/pkg/i18n/langimpl"

var (
	ErrInvalidFileType               = ErrParams.NewWithSubId(langimpl.SubIdInvalidFileType)
	ErrThirdPartyServiceNameNotMatch = ErrParams.NewWithSubId(langimpl.SubIdThirdPartyServiceNameNotMatch)
	ErrEmailCodeNeedSendFirst        = ErrParams.NewWithSubId(langimpl.SubIdEmailCodeNeedSendFirst)
	ErrWriteFileFailed               = ErrParams.NewWithSubId(langimpl.SubIdWriteFileFailed)
)
