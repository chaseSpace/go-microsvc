package currency

import (
	"microsvc/enums"
	"microsvc/infra/svccli/rpcext"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/currencypb"
	"microsvc/service/currency/deploy"
	"microsvc/test/tbase"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	tbase.TearUp(enums.SvcCurrency, deploy.CurrencyConf)
}

func TestGetGoldAccount(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Currency().GetGoldAccount(tbase.TestCallCtx, &currencypb.GetGoldAccountReq{
		Base: tbase.TestBaseExtReq,
	})
}

func TestGetGoldTxLog(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Currency().GetGoldTxLog(tbase.TestCallCtx, &currencypb.GetGoldTxLogReq{
		Base:       tbase.TestBaseExtReq,
		OrderField: "created_at",
		OrderType:  commonpb.OrderType_OT_Desc,
		YearMonth:  "202408",
		Page: &commonpb.PageArgs{
			Pn: 2,
			Ps: 2,
		},
	})
}

func TestTestGoldTx(t *testing.T) {
	defer tbase.TearDown()

	rpcext.Currency().TestGoldTx(tbase.TestCallCtx, &currencypb.TestGoldTxReq{
		Base:   tbase.TestBaseExtReq,
		Uid:    1,
		Delta:  -1001,
		TxType: currencypb.GoldTxType_GSTT_Recharge,
		Remark: "充值",
	})
}

// 并发测试
// -- 观察数据库是否报订单号重复的错误，余额是否正确
func TestConcurrencyTestGoldTx(t *testing.T) {
	defer tbase.TearDown()

	r, err := rpcext.Currency().GetGoldAccount(tbase.TestCallCtx, &currencypb.GetGoldAccountReq{
		Base: tbase.TestBaseExtReq,
	})
	if err != nil {
		t.Fatal(err)
	}

	st := time.Now()
	N := 200 // 过高会导致mysql连接数超限而阻塞/报错
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := tbase.NewTestCallCtxWithTimeout(time.Second*3, 1)
			defer cancel()
			_, err = rpcext.Currency().TestGoldTx(ctx, &currencypb.TestGoldTxReq{
				Base:   tbase.TestBaseExtReq,
				Uid:    1,
				Delta:  1,
				TxType: currencypb.GoldTxType_GSTT_Recharge,
				Remark: "哈哈",
			})
			if err != nil {
				t.Error("ERROR", err)
			}
		}()
	}

	wg.Wait()

	t.Logf("avg time cost: %dms\n", time.Since(st).Milliseconds()/int64(N))
	r2, _ := rpcext.Currency().GetGoldAccount(tbase.TestCallCtx, &currencypb.GetGoldAccountReq{
		Base: tbase.TestBaseExtReq,
	})
	assert.Equal(t, r.Balance, r2.Balance-int64(N))
}
