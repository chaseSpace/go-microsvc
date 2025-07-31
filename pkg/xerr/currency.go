package xerr

import "microsvc/pkg/i18n/langimpl"

var (
	ErrTxFromAndToCannotBeSame = ErrParams.NewWithSubId(langimpl.SubIdTxFromAndToCannotBeSame)

	ErrTxAmountShouldNotBeZero = ErrParams.NewWithSubId(langimpl.SubIdTxAmountShouldNotBeZero)
	ErrTxRemarkTooLong         = ErrParams.NewWithSubId(langimpl.SubIdTxRemarkTooLong)
	ErrTxBalanceNotEnough      = ErrParams.NewWithSubId(langimpl.SubIdTxBalanceNotEnough)
	ErrTxInvalidTxType         = ErrParams.NewWithSubId(langimpl.SubIdTxInvalidTxType)
	ErrTxInvalidSingleTxType   = ErrParams.NewWithSubId(langimpl.SubIdTxInvalidSingleTxType)
)
