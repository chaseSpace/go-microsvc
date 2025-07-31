package tongyi

import (
	"encoding/json"
	"microsvc/pkg/xerr"
	"time"

	"github.com/parnurzeal/gorequest"
)

const baseUrl = "https://dashscope.aliyuncs.com/compatible-mode"

type client struct {
	apiKey string
	agent  *gorequest.SuperAgent
}

var cli client

func Init(apiKey string) {
	if apiKey == "" {
		panic("tongyi api key is empty")
	}
	cli.apiKey = apiKey
	cli.agent = gorequest.New().AppendHeader("Authorization", "Bearer "+apiKey).Timeout(time.Second * 10)
	cli.agent.DoNotClearSuperAgent = true
}

func doRequest(endpoint string, req interface{}, res interface{}) error {
	r, _buf, errs := cli.agent.Post(baseUrl + endpoint).SendStruct(req).EndStruct(&res)
	if len(errs) > 0 {
		return xerr.JoinErrors(errs...)
	}
	if r.StatusCode != 200 {
		e := new(ErrorResponse)
		_ = json.Unmarshal(_buf, e)
		return e
	}
	return nil
}

func ChatCompletions(req *ChatCompletionReq) (res *ChatCompletionRes, err error) {
	return res, doRequest("/v1/chat/completions", req, &res)
}
