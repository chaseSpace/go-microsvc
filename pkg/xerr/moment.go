package xerr

import "microsvc/pkg/i18n/langimpl"

var (
	ErrMomentTextTooLong  = ErrParams.NewWithSubId(langimpl.SubIdMomentTextTooLong)
	ErrMomentTypeNotFound = ErrParams.NewWithSubId(langimpl.SubIdMomentTypeNotFound)
	ErrMomentNotFound     = ErrParams.NewWithSubId(langimpl.SubIdMomentNotFound)
	ErrCommentNotFound    = ErrParams.NewWithSubId(langimpl.SubIdCommentNotFound)
	ErrCommentTextTooLong = ErrParams.NewWithSubId(langimpl.SubIdCommentTextTooLong)
)
