package abstract

import (
	"fmt"
	"microsvc/pkg/xlog"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type JobAPI interface {
	cron.Job
	RunStateApi
	Name() string
	RunID() string
	Run2() error
}

func ComputeTimeCost(j cron.Job) cron.Job {
	return cron.FuncJob(func() {
		j2 := j.(JobAPI)
		uniqID := j2.RunID()

		allowParallel := j2.AllowParallel()
		if j2.IsRunning() {
			if !allowParallel {
				xlog.Warn("crontask 尚未停止", zap.String("name", j2.Name()), zap.String("uniqID", uniqID),
					zap.String("cost", j2.CostTime().String()))
				return
			}
		}
		if !j2.Start(uniqID, allowParallel) {
			return
		}
		defer j2.Stop(uniqID)

		xlog.Info("crontask 开始运行", zap.String("name", j2.Name()), zap.String("uniqID", uniqID),
			zap.String("DaemonsDetail", j2.DaemonsDetail()))

		var err error
		defer func() {
			xlog.Info("crontask 结束运行", zap.String("name", j2.Name()), zap.String("uniqid", uniqID),
				zap.String("cost", j2.CostTime().String()), zap.Error(err))
		}()
		err = j2.Run2()
	})
}

type RunState struct {
	isRunning atomic.Bool
	startAt   time.Time
	container sync.Map
}

// Start 运行前调用，在并发调用时只有一个调用会成功（理论上不会出现失败）
func (s *RunState) Start(uniqID string, allowParallel bool) bool {
	ok := s.isRunning.CompareAndSwap(false, true)
	if !ok {
		if allowParallel {
			s.container.Store(uniqID, time.Now())
			return true
		}
		return ok
	}
	s.startAt = time.Now()
	return ok
}
func (s *RunState) Stop(uniqID string) {
	if _, ok := s.container.Load(uniqID); ok {
		s.container.Delete(uniqID)
		return
	}
	s.isRunning.Store(false)
	s.startAt = time.Time{}
}
func (s *RunState) Daemons() int {
	var ct int
	s.container.Range(func(key, value any) bool {
		ct++
		return true
	})
	return ct
}

func (s *RunState) DaemonsDetail() string {
	type Temp struct {
		UniqID string
		Dur    time.Duration
	}
	var items []*Temp

	s.container.Range(func(key, value any) bool {
		items = append(items, &Temp{
			UniqID: key.(string),
			Dur:    time.Since(value.(time.Time)),
		})
		return true
	})
	sort.Slice(items, func(i, j int) bool { return items[i].Dur > items[j].Dur })
	var desc []string
	for _, item := range items {
		desc = append(desc, fmt.Sprintf(`【%s: %s】`, item.UniqID, item.Dur))
	}
	return fmt.Sprintf(`daemons: %d `, s.Daemons()) + strings.Join(desc, " - ")
}

func (s *RunState) CostTime() time.Duration {
	return time.Since(s.startAt)
}

func (s *RunState) IsRunning() bool {
	return s.isRunning.Load()
}
func (s *RunState) AllowParallel() bool {
	return false
}

type RunStateApi interface {
	Start(uniqID string, allowParallel bool) bool
	Stop(uniqID string)
	Daemons() int
	DaemonsDetail() string
	CostTime() time.Duration
	IsRunning() bool
	AllowParallel() bool
}
