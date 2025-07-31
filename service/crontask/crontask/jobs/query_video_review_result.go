package jobs

import (
	"context"
	"microsvc/infra/svccli/rpc"
	"microsvc/model/svc/admin"
	"microsvc/pkg/xlog"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/momentpb"
	"microsvc/protocol/svc/thirdpartypb"
	"microsvc/service/crontask/crontask/abstract"
	"microsvc/service/crontask/dao"
	"microsvc/util"
	"microsvc/util/urand"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type QueryVideoReviewResult struct {
	cron.Job
	abstract.RunState
	*zap.Logger
}

func (t *QueryVideoReviewResult) Name() string {
	return "review.QueryVideoReviewRet"
}

func (t *QueryVideoReviewResult) RunID() string {
	return urand.Strings(4)
}

// Run2 查询视频审核结果
// - 1. 先从审核表中查出已提交审核的记录
// - 2. 使用记录中的唯一id，去三方查询审核结果
// - 3. 将结果通知到对应服务
func (t *QueryVideoReviewResult) Run2() (err error) {
	t.Logger = xlog.WithFields(zap.String("CronTask", t.Name()))

	// 预估耗时：5分钟
	util.RunTaskWithCtxTimeout(time.Minute*5, func(ctx context.Context) {
		err = t.RunCore(ctx)
	})
	return
}

func (t *QueryVideoReviewResult) RunCore(ctx context.Context) (err error) {
	limit := 100
	for {
		var list []*admin.ReviewVideo
		list, err = dao.AdminReviewDao.ListReviewVideoPendingRecords(ctx, limit)
		if err != nil {
			return err
		}

		t.Debug("ListReviewVideoPendingRecords", zap.Int("len", len(list)))

		if len(list) == 0 {
			return
		}

		var okCnt, errs int
		for i, v := range list {
			if err = t.__loopLogic(ctx, i, v); err != nil {
				errs++
				t.Error("RunCore in loop", zap.Error(err), zap.Int("seq", i), zap.Any("item", v))

				// 不管什么err，都更新记录的失败次数（是为了避免某些记录存在问题，导致定时任务无限循环）
				err = dao.AdminReviewDao.UpdateReviewVideoFails(ctx, v.Id)
				if err != nil {
					t.Error("UpdateReviewVideoFails", zap.Int("seq", i), zap.Error(err))
					return err // 严重错误，直接返回
				}
				continue
			}
			okCnt++
		}

		t.Info("RunCore Single Loop END", zap.Int("okCnt", okCnt), zap.Int("errs", errs))
		if len(list) < limit {
			return
		}
		time.Sleep(time.Second)
	}
}

func (t *QueryVideoReviewResult) __loopLogic(ctx context.Context, i int, item *admin.ReviewVideo) (err error) {
	res, err := rpc.Thirdparty().QueryVideoReviewResult(ctx, &thirdpartypb.QueryVideoReviewResultReq{
		ThName: item.ThName,
		Ext: &thirdpartypb.ReviewParamsExt{
			UniqReqId: &thirdpartypb.ReviewParamsExt_UniqReqId{Val: item.ThTaskId},
		}})
	if err != nil {
		return
	}
	if res.Status == commonpb.AIReviewStatus_ARS_Pending {
		t.Info("__loopLogic got pending record, SKIP", zap.Int("seq", i), zap.Any("item", item))
		return
	}

	rs := __convertAIReviewStatus(res.Status)
	if rs == 0 {
		return errors.New("不支持的AIReviewStatus: " + res.Status.String())
	}
	// 更新记录
	err = dao.AdminReviewDao.UpdateReviewVideoStatus(ctx, item.Id, rs)
	if err != nil {
		return
	}

	switch res.Status {
	case commonpb.AIReviewStatus_ARS_Reject:
	default:
		// 除了拒绝的其他状态，不会通知目标服务，还要经过人审
		return
	}

	// 通知目标服务
	switch item.BizType {
	case commonpb.BizType_RBT_Moment:
		ms := __convertAIReviewStatusToMomentS(res.Status)
		_, err = rpc.Moment().UpdateReviewStatus(ctx, &momentpb.UpdateReviewStatusReq{
			Uid:    item.UID,
			Mid:    item.BizUniqId,
			Status: ms,
		})
	default:
		return errors.New("不支持的BizType: " + item.BizType.String())
	}
	return
}

func __convertAIReviewStatus(status commonpb.AIReviewStatus) commonpb.ReviewStatus {
	return map[commonpb.AIReviewStatus]commonpb.ReviewStatus{
		commonpb.AIReviewStatus_ARS_Pass:   commonpb.ReviewStatus_RS_AIPass,
		commonpb.AIReviewStatus_ARS_Reject: commonpb.ReviewStatus_RS_AIReject,
		commonpb.AIReviewStatus_ARS_Review: commonpb.ReviewStatus_RS_Manual,
	}[status]
}

func __convertAIReviewStatusToMomentS(status commonpb.AIReviewStatus) momentpb.ReviewStatus {
	return map[commonpb.AIReviewStatus]momentpb.ReviewStatus{
		commonpb.AIReviewStatus_ARS_Pass:   momentpb.ReviewStatus_RS_Pass,
		commonpb.AIReviewStatus_ARS_Reject: momentpb.ReviewStatus_RS_Reject,
	}[status]
}
