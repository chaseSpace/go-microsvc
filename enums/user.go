package enums

import "microsvc/protocol/svc/commonpb"

type Sex int32

const (
	SexNoSet Sex = iota
	SexMale
	SexFemale
	SexMax
)

func (s Sex) Int32() int32 {
	return int32(s)
}

func (s Sex) ToPB() commonpb.Sex {
	return commonpb.Sex(s)
}

func (s Sex) IsInvalid() bool {
	return !(s > SexNoSet && s < SexMax)
}

type UserWxType int8

const (
	UserWxTypeApp UserWxType = iota
	UserWxTypeMini
	UserWxTypeOfficialAccount
)
