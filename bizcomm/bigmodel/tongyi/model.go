package tongyi

import (
	"fmt"
)

type ErrorResponse struct {
	Errors ErrorDetail `json:"error"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("Error: %s, Type: %s", e.Errors.Message, e.Errors.Type)
}

type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// --------- Req & Res ----------------

type ImageURL struct {
	Type     string `json:"type"`
	ImageURL struct {
		URL string `json:"url"`
	} `json:"image_url"`
}

type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Message struct {
	Role    string        `json:"role"`
	Content []interface{} `json:"content"`
}

type ChatCompletionReq struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}
type ChatCompletionRes struct {
	Choices           []Choice    `json:"choices"`
	Object            string      `json:"object"`
	Usage             Usage       `json:"usage"`
	Created           int64       `json:"created"`
	SystemFingerprint interface{} `json:"system_fingerprint"` // 使用指针类型以处理可能为 null 的情况
	Model             string      `json:"model"`
	ID                string      `json:"id"`
}

type CCResMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type Choice struct {
	Message      CCResMessage `json:"message"`
	FinishReason string       `json:"finish_reason"`
	Index        int          `json:"index"`
	Logprobs     interface{}  `json:"logprobs"` // 使用指针类型以处理可能为 null 的情况
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ---------------- Req Builder --------------------

type chatCmplReqBuilder struct {
	model string
}

var ChatCmplReqBuilder = &chatCmplReqBuilder{}

func (c *chatCmplReqBuilder) UseQwenVlMax() *chatCmplReqBuilder {
	c.model = "qwen-vl-max"
	return c
}

func (c *chatCmplReqBuilder) SummaryViaImageBufAndText(imageB64Buf []byte, text, prompt string) *ChatCompletionReq {
	if c.model == "" {
		panic("model not set")
	}
	return &ChatCompletionReq{
		Model: c.model,
		Messages: []Message{
			{
				Role: "user",
				Content: []interface{}{
					ImageURL{
						Type: "image_url",
						ImageURL: struct {
							URL string `json:"url"`
						}{
							URL: fmt.Sprintf("data:image/jpeg;base64,%s", imageB64Buf),
						},
					},
					Text{
						Type: "text",
						Text: "文本参数：" + text,
					},
					Text{
						Type: "text",
						Text: prompt,
					},
				},
			},
		},
	}
}

func (c *chatCmplReqBuilder) SummaryViaImageUrlAndText(imageUrl string, text, prompt string) *ChatCompletionReq {
	if c.model == "" {
		panic("model not set")
	}
	return &ChatCompletionReq{
		Model: c.model,
		Messages: []Message{
			{
				Role: "user",
				Content: []interface{}{
					ImageURL{
						Type: "image_url",
						ImageURL: struct {
							URL string `json:"url"`
						}{
							//URL: fmt.Sprintf("data:image/jpeg;base64,%s", imageBuf.Bytes()),
							URL: imageUrl,
						},
					},
					Text{
						Type: "text",
						Text: "文本参数：" + text,
					},
					Text{
						Type: "text",
						Text: prompt,
					},
				},
			},
		},
	}
}
