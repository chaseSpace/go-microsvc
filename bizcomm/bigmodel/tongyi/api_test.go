package tongyi

import (
	"bytes"
	"encoding/base64"
	"io"
	"microsvc/util/uregex"
	"os"
	"testing"

	"github.com/k0kubun/pp/v3"
)

func getImgBase64(name string) (buf *bytes.Buffer, err error) {
	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}

	b, _ := io.ReadAll(f)
	buf = &bytes.Buffer{}
	_, err = base64.NewEncoder(base64.StdEncoding, buf).Write(b)
	return
}
func TestChatCompletions(t *testing.T) {
	Init("sk-04a356f1d43644dca89c659fe2070ac9")

	url := "https://free.wzznft.com/i/2024/09/13/h5q1a2.jpg"
	//url := "https://static.zixiaoyun.com/i/2023/07/25/u5x13.png"
	text := "xxx"
	prompt := `提取上述图片和文本参数中的信息，判断是否一个酒吧活动预告，若不是返回0，若是则返回一个不含换行的JSON字符串
{"date_info":string,"Topic":string, "is_bar_activity":bool}，不要包含多余的文本`
	res, err := ChatCompletions(ChatCmplReqBuilder.UseQwenVlMax().SummaryViaImageUrlAndText(url, text, prompt))
	if err != nil {
		t.Fatal(err)
	}
	pp.Println(res)

	if len(res.Choices) == 0 {
		t.Fatal("no choices")
	}
	type result struct {
		DateInfo string `json:"date_info"`
		Topic    string `json:"topic"`
	}
	var r result
	_ = uregex.ExtractJson(res.Choices[0].Message.Content, &r)
	pp.Println(r)
}
