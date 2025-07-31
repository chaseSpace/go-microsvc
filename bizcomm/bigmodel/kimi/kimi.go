package kimi

import (
	"context"
	"microsvc/pkg/xlog"
	"microsvc/xvendor/moonshot_sdk"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Moonshot struct {
	baseUrl string
	key     string
	client  *http.Client
}

func (m Moonshot) BaseUrl() string      { return m.baseUrl }
func (m Moonshot) Key() string          { return m.key }
func (m Moonshot) Client() *http.Client { return m.client }

func (m Moonshot) Log(ctx context.Context, caller string, request *http.Request, response *http.Response, elapse time.Duration) {
	xlog.Debug("Kimi-api", zap.String("caller", caller), zap.String("path", request.URL.Path), zap.Duration("elapse", elapse))
}

var once sync.Once
var client moonshot_sdk.Client[Moonshot]

func Init(apiKey string) {
	if apiKey == "" {
		panic("kimi api key is empty")
	}
	once.Do(func() {
		client = moonshot_sdk.NewClient[Moonshot](Moonshot{
			baseUrl: "https://api.moonshot.cn/v1",
			key:     apiKey,
			client:  http.DefaultClient,
		})
	})

	balance, err := client.CheckBalance(context.TODO())
	if err != nil {
		panic(err)
	}
	xlog.Info("Kimi-Init", zap.Any("balance", balance.Data))
}

func Client() moonshot_sdk.Client[Moonshot] {
	return client
}
