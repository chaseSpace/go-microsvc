package xnotify

type Scene string

const (
	SceneServerException Scene = "server_exception"
	SceneMqException     Scene = "mq_exception"
	SceneOrderException  Scene = "order_exception"
)
