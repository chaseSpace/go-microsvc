package xerr

import "microsvc/pkg/i18n/langimpl"

var (
	ErrPasswdFormat           = ErrParams.NewWithSubId(langimpl.SubIdPasswdFormat)
	ErrPasswdNoChange         = ErrParams.NewWithSubId(langimpl.SubIdPasswdNoChange)
	ErrInvalidPhoneNo         = ErrParams.NewWithSubId(langimpl.SubIdInvalidPhoneNo)
	ErrInvalidVerifyCode      = ErrParams.NewWithSubId(langimpl.SubIdInvalidVerifyCode)
	ErrInvalidLenPhoneNo      = ErrParams.NewWithSubId(langimpl.SubIdInvalidLenPhoneNo)
	ErrNotSupportedPhoneArea  = ErrParams.NewWithSubId(langimpl.SubIdNotSupportedPhoneArea)
	ErrChangePasswdFrequently = ErrParams.NewWithSubId(langimpl.SubIdChangePasswdFrequently)
	ErrUnRegisteredPhone      = ErrParams.NewWithSubId(langimpl.SubIdUnRegisteredPhone)
	ErrUnSupportedSignInType  = ErrParams.NewWithSubId(langimpl.SubIdUnSupportedSignInType)
	ErrIncorrectPassword      = ErrParams.NewWithSubId(langimpl.SubIdIncorrectPassword)
	ErrSignInBanned           = ErrParams.NewWithSubId(langimpl.SubIdSignInBanned)
	ErrAccountBanned          = ErrParams.NewWithSubId(langimpl.SubIdAccountBanned)
	ErrSignInFailed           = ErrParams.NewWithSubId(langimpl.SubIdSignInFailed)
	ErrAccountAlreadyExists   = ErrParams.NewWithSubId(langimpl.SubIdAccountAlreadyExists)
	ErrLoginFrequently        = ErrParams.NewWithSubId(langimpl.SubIdLoginFrequently)
	ErrLoginFrequentlyLong    = ErrParams.NewWithSubId(langimpl.SubIdLoginFrequentlyLong)
	ErrPasswordNotSetOnLogin  = ErrParams.NewWithSubId(langimpl.SubIdPasswordNotSetOnLogin)
	ErrPasswordNotGiven       = ErrParams.NewWithSubId(langimpl.SubIdPasswordNotGiven)
)
