package handler

import (
	"microsvc/protocol/svc/friendpb"
)

type intCtrl struct {
}

var IntCtrl friendpb.FriendIntServer = new(intCtrl)
