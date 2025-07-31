package base

import (
	"microsvc/bizcomm/commuser"
)

// 一些子服务需要全局使用的资源在这里初始化

var (
	PhoneTool = commuser.PhoneTool
)

func MustInit() {

}
