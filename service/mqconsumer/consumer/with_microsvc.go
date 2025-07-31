package consumer

import (
	"context"
	"microsvc/bizcomm/mq"
	"microsvc/consts"
	"microsvc/infra/xmq"
	"microsvc/infra/xmq/define"
	"microsvc/model/svc/micro_svc"
	"microsvc/util/db"
	"time"

	"gorm.io/gorm"
)

type withMicroSvcT struct {
}

var consumerWithMicroSvc = new(withMicroSvcT)

func (w *withMicroSvcT) Init() {
	go w.ConsumeAPICallLog()
}

func (w *withMicroSvcT) ConsumerName() string {
	return "MicroSvc"
}

func (w *withMicroSvcT) ConsumeAPICallLog() {
	topic := consts.TopicAPICallLog
	xmq.Consume[*mq.MsgAPICallLog](topic, func(ctx context.Context, _msg *mq.MsgAPICallLog) error {
		// 受限于GO泛型规则，这里似乎没有更好的写法
		// - 注意，这里分发给协程的msg要使用值类型，避免潜在的并发修改问题
		msg := *_msg

		ctx = context.TODO()
		return consumerAPICallLogV.Archive(ctx, msg)
	}, define.ConsumeExtraArg{ConsumeGroupId: consts.CGDefault})
}

// ----------------- 优雅的分割线（下面定义不同topic的多个消费方法） -----------------

type consumerAPICallLog struct {
}

var consumerAPICallLogV consumerAPICallLog

func (consumerAPICallLog) Archive(ctx context.Context, msg mq.MsgAPICallLog) error {
	callLog := &micro_svc.APICallLog{
		CreatedAt: time.UnixMilli(msg.MsTimestamp),
	}
	callLog.SetInner(msg.APICallLogBody)
	callLog.SetSuffix(callLog.CreatedAt.Format("200601"))
	err := db.NewTableHelper(micro_svc.Q.WithContext(ctx), callLog.DLLSql()).AutoCreateTable(func(tx *gorm.DB) error {
		return tx.Table(callLog.TableName()).Create(callLog).Error
	})
	if err != nil {
		return err
	}
	return nil
}
