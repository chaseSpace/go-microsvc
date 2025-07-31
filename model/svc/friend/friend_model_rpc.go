package friend

import (
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/friendpb"
	"time"
)

type FriendRPC struct {
	*Friend
	UserPB *commonpb.User
}

func (f *FriendRPC) ToPB() *friendpb.Friend {
	return &friendpb.Friend{
		User:         f.UserPB,
		CreatedAtTs:  f.CreatedAt.Unix(),
		CreatedAtStr: f.CreatedAt.Format(time.DateTime),
		Intimacy:     f.Intimacy,
	}
}

type BlockRPC struct {
	*Block
	UserPB *commonpb.User
}

func (b *BlockRPC) ToPB() *friendpb.BlockUser {
	return &friendpb.BlockUser{
		User:         b.UserPB,
		CreatedAtTs:  b.CreatedAt.Unix(),
		CreatedAtStr: b.CreatedAt.Format(time.DateTime),
	}
}

type VisitorRPC struct {
	*Visitor
	UserPB *commonpb.User
}

func (v *VisitorRPC) ToPB() *friendpb.Visitor {
	return &friendpb.Visitor{
		User:         v.UserPB,
		CreatedAtTs:  v.CreatedAt.Unix(),
		CreatedAtStr: v.CreatedAt.Format(time.DateTime),
		Date:         v.Date,
		Desc:         "",
		ReplaceElem:  nil,
	}
}
