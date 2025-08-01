package xlog

import (
	"microsvc/deploy"
	"microsvc/util"
	"microsvc/util/utime"
	"os"

	"github.com/k0kubun/pp/v3"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var xlogger *zap.Logger

var svc string

func Init(cc *deploy.XConfig) {
	svc = cc.Svc.Name()
	var lv = zapcore.DebugLevel
	switch cc.GetSvcConf().GetLogLevel() {
	case "info":
		lv = zapcore.InfoLevel
	case "warn":
		lv = zapcore.WarnLevel
	case "error":
		lv = zapcore.ErrorLevel
	}
	writer := zapcore.AddSync(os.Stdout) // 写stdout，再用容器收集日志
	core := zapcore.NewCore(getEncoder(), writer, lv)
	xlogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

func Stop() {
	_ = xlogger.Sync()
}

func getEncoder() zapcore.Encoder {
	var ec = zap.NewProductionEncoderConfig()
	customLevelEncoder := func(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString("x-" + level.String()) // x-error  x-info 更容易与调用者打印的error字面量区分开
	}
	ec.EncodeLevel = customLevelEncoder
	ec.EncodeTime = zapcore.TimeEncoderOfLayout(utime.DateTimeMs)
	ec.LevelKey = "LEVEL"
	ec.TimeKey = "TIME"
	ec.CallerKey = "CALLER"
	ec.MessageKey = "MSG"
	ec.StacktraceKey = "STACK"
	ec.LineEnding = "\n\n" // 大量日志时有助于观察
	//return zapcore.NewConsoleEncoder(ec) // 按行打印
	return zapcore.NewJSONEncoder(ec)
}

// --------------------------------

func appendFields(fields *[]zapcore.Field) {
	*fields = append(*fields, zap.String("SERVICE", svc))
}

func WithFields(fields ...zapcore.Field) *zap.Logger {
	appendFields(&fields)
	return xlogger.With(fields...)
}

func Debug(msg string, fields ...zapcore.Field) {
	appendFields(&fields)
	xlogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zapcore.Field) {
	appendFields(&fields)
	xlogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zapcore.Field) {
	_, _ = util.WaringPP.Println("***** Warn: " + msg)
	_, _ = pp.Printf(util.FormatStack(3) + "\n")
	appendFields(&fields)
	xlogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	_, _ = pp.Println("***** Error: " + msg)
	_, _ = pp.Printf(util.FormatStack(3) + "\n")
	appendFields(&fields)
	xlogger.Error(msg, fields...)
}

func ErrorNoStack(msg string, fields ...zapcore.Field) {
	appendFields(&fields)
	xlogger.Error(msg, fields...)
}

func Panic(msg string, fields ...zapcore.Field) {
	appendFields(&fields)
	xlogger.Panic(msg, fields...)
}

func DPanic(msg string, fields ...zapcore.Field) {
	_, _ = pp.Println("***** Panic: " + msg)
	_, _ = pp.Printf(util.FormatStack(5) + "\n")
	appendFields(&fields)
	xlogger.DPanic(msg, fields...)
}

// 这里不需要DPanic函数，因为Panic够用，且我们的grpc中间件会捕获panic，并封装包含panic信息的Response
