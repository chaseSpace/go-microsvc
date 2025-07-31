package xnotify

import (
	"context"
	"github.com/blinkbean/dingtalk"
	"github.com/stretchr/testify/assert"
	"microsvc/deploy"
	"os"
	"testing"
)

func init() {
	Init(&deploy.XConfig{
		ExternalNotify: deploy.ExternalNotify{
			Dingtalk: map[string]*struct {
				RobotToken string `mapstructure:"robot_token"`
				Secret     string `mapstructure:"secret"`
			}{
				"server_exception": {
					RobotToken: os.Getenv("DINGTALK_TOKEN"),
					Secret:     os.Getenv("DINGTALK_SECRET"),
				},
			},
		},
	})
}

func TestNotifyDingtalkText(t *testing.T) {
	err := NotifyDingtalkText(context.TODO(), SceneServerException, NotifyArgs{
		Title:   "站到！",
		Content: "Hi，我是一名重庆靓仔！",
	})
	assert.Nil(t, err)
}

func TestNotifyDingtalkMD(t *testing.T) {
	err := NotifyDingtalkMD(context.TODO(), SceneServerException, NotifyArgs{
		Title: "颜色测试",
		//Content: "Content",
		MarkdownLines: []*MdLine{
			{
				ContentWithPlaceHolder: "失败：$$ 同行不同色 $$",
				MarkType:               dingtalk.RED,
			},
			{
				ContentWithPlaceHolder: "---",
				MarkType:               dingtalk.N,
			},
			{
				ContentWithPlaceHolder: "金色",
				MarkType:               dingtalk.GOLD,
			},
			{
				ContentWithPlaceHolder: "成功",
				MarkType:               dingtalk.GREEN,
			},
		},
	})
	assert.Nil(t, err)
}
