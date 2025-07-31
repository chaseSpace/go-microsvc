package xerr

import (
	"microsvc/pkg/i18n/langimpl"
)

var (
	ErrGiftTxFromAndToCannotBeSame = ErrParams.NewWithSubId(langimpl.SubIdGiftTxFromAndToCannotBeSame)

	ErrGiftTxAmountMustBePositive       = ErrParams.NewWithSubId(langimpl.SubIdGiftTxAmountMustBePositive)
	ErrGiftTxAmountMustBePositiveOnSent = ErrParams.NewWithSubId(langimpl.SubIdGiftTxAmountMustBePositiveOnSent)
	ErrGiftTxInvalidTxType              = ErrParams.NewWithSubId(langimpl.SubIdGiftTxInvalidTxType)
	ErrGiftTxInvalidTxScene             = ErrParams.NewWithSubId(langimpl.SubIdGiftTxInvalidTxScene)
	ErrGiftTxInvalidFirstPersonTxType   = ErrParams.NewWithSubId(langimpl.SubIdGiftTxInvalidFirstPersonTxType)
	ErrGiftTxTypeConvertFailed          = ErrParams.NewWithSubId(langimpl.SubIdGiftTxTypeConvertFailed)
	ErrGiftTxBalanceNotEnough           = ErrParams.NewWithSubId(langimpl.SubIdGiftTxBalanceNotEnough)
	ErrGiftNotFound                     = ErrParams.NewWithSubId(langimpl.SubIdGiftNotFound)
)
