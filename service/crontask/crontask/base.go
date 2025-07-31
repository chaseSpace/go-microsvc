package crontask

import (
	"microsvc/enums"
	"microsvc/pkg/xlog"
	"microsvc/service/crontask/crontask/abstract"
	"microsvc/service/crontask/crontask/jobs"
	"microsvc/util"

	"github.com/spf13/cobra"

	"github.com/robfig/cron/v3"
)

var registerFlag = 0

func addJob(spec string, job abstract.JobAPI) {
	_, err := cc.AddJob(spec, job)
	util.AssertNilErr(err)
	registerFlag = 1
}

var cc = cron.New(cron.WithSeconds(), cron.WithChain(abstract.ComputeTimeCost))

func MustInit() {
	//err := new(jobs.InstagramSpider).Run2()
	//util.AssertNilErr(err)

	rootCmd.AddCommand(__QueryVideoReviewResult)

	util.AssertNilErr(rootCmd.Execute())

	if registerFlag == 0 {
		panic("no job registered")
	}
	go cc.Run()
}

var rootCmd = &cobra.Command{
	Use: enums.SvcCrontask.Name(),
}

// 下面是子命令，一个进程只能启动一个任务，独立管理

var __QueryVideoReviewResult = &cobra.Command{
	Use: "QueryVideoReviewResult",
	Run: func(cmd *cobra.Command, args []string) {
		addJob("*/30 * * * * *", new(jobs.QueryVideoReviewResult)) // 30s1次
		xlog.Info("+++++++++ QueryVideoReviewResult registered +++++++++")
	},
}
