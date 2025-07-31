package xnotify

import (
	"context"
	"fmt"
	"github.com/blinkbean/dingtalk"
	"github.com/k0kubun/pp/v3"
	"microsvc/deploy"
)

type NotifyArgs struct {
	Title, Content string

	MarkdownLines []*MdLine // 当Content非空时，该字段无效
}

type MdLine struct {
	ContentWithPlaceHolder string
	MarkType               dingtalk.MarkType
}

var (
	dingtalkCliMap = make(map[Scene]*dingtalk.DingTalk)
)

func Init(cc *deploy.XConfig) {
	for scene, c := range cc.ExternalNotify.Dingtalk {
		dingtalkCliMap[Scene(scene)] = dingtalk.InitDingTalkWithSecret(c.Token, c.Secret)
	}
}

func getCli(scene Scene) (*dingtalk.DingTalk, error) {
	cli := dingtalkCliMap[scene]
	if cli == nil {
		_, _ = pp.Printf("dingtalk cli is nil, scene [%s], call Init() firstly in main.go\n", scene)
		return nil, fmt.Errorf("dingtalk cli is nil, scene [%s]", scene)
	}
	return cli, nil
}

func NotifyDingtalkText(ctx context.Context, scene Scene, args NotifyArgs) error {
	cli, err := getCli(scene)
	if err != nil {
		return err
	}

	ct := args.Content
	err = cli.SendTextMessageWithCtx(ctx, "标题："+args.Title+"\n正文："+ct)
	return err
}

func NotifyDingtalkMD(ctx context.Context, scene Scene, args NotifyArgs) error {
	cli, err := getCli(scene)
	if err != nil {
		return err
	}

	ss := dingtalk.DingMap()
	ss.Set(args.Title, dingtalk.H2)
	if args.Content != "" {
		ss.Set(args.Content, dingtalk.N)
	} else {
		for _, line := range args.MarkdownLines {
			ss.Set(line.ContentWithPlaceHolder, line.MarkType)
		}
	}
	err = cli.SendMarkDownMessageBySliceWithCtx(ctx, args.Title, ss.Slice())
	return err
}
