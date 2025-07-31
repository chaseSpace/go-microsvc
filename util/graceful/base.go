package graceful

import (
	"microsvc/pkg/xlog"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

var sigChan = make(chan os.Signal, 1)

func SetupSignal() {
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
}

const logPrefix = "****** graceful ****** "

var stopFuncSlice []func()

func AddStopFunc(f func()) {
	stopFuncSlice = append(stopFuncSlice, f)
}

func OnExit() {
	stopAll() // case 2: backgroundSvc exited normally,  or signal received
	if err := recover(); err != nil {
		xlog.Panic(logPrefix+"server exited (goroutine panic)", zap.Any("err", err))
	}
	xlog.Debug(logPrefix + "server exited")
	xlog.Stop()
}

func Stop() {
	sigChan <- syscall.SIGTERM
}

func stopAll() {
	for _, stopF := range stopFuncSlice {
		stopF()
	}
}

func Register(f func()) {
	go func() {
		f()
		Stop()
	}()
}

func Run() {
	reason := ""
	select {
	case <-sigChan:
		reason = "(signal)"
	}
	xlog.Info(logPrefix + "server ready to exit" + reason)
}
