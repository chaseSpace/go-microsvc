package logic

type intCtrl struct{}

var Int = intCtrl{} // 暴露struct而不是interface，方便IDE跳转
