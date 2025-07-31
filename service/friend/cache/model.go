package cache

import (
	"microsvc/model/svc/friend"
	"microsvc/protocol/svc/commonpb"
)

type friendListT struct {
	List  []*friend.Friend
	Total int64
}

type visitorListT struct {
	List             []*friend.Visitor
	Total            int64
	VisitorsTotal    *commonpb.CounterInt64
	VisitorsRepeated *commonpb.CounterInt64
}
