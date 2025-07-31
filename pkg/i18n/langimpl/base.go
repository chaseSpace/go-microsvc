package langimpl

import "microsvc/protocol/svc/commonpb"

type (
	Code  int32
	Scene string
	SubId string
)

func (c Code) Int32() int32 {
	return int32(c)
}

const (
	SceneError Scene = "error"
)

type MsgMap map[Code]map[SubId]string

// LangAPI 所有语言要实现的接口
type LangAPI interface {
	Lang() commonpb.Lang
	Translate(scene Scene, id Code, subId SubId) string
}
