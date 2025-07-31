package xerr

import "microsvc/pkg/i18n/langimpl"

var (
	ErrFriendAlreadyFollow    = ErrParams.NewWithSubId(langimpl.SubIdFriendAlreadyFollow)
	ErrFriendCountUpToMax     = ErrParams.NewWithSubId(langimpl.SubIdFriendCountUpToMax)
	ErrFriendPeerCountUpToMax = ErrParams.NewWithSubId(langimpl.SubIdFriendPeerCountUpToMax)
)
