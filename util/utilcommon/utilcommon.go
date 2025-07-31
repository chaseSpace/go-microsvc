package utilcommon

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

func PrintlnStackMsg(msg string, args ...interface{}) {
	_, file, line, ok := runtime.Caller(1) // 跳过PrintlnStackErr
	if !ok {
		file = "???"
		line = 0
	}
	formattedMsg := fmt.Sprintf(msg, args...)
	stack := TraceStack(3)
	log.Printf("%s:%d %s\n---StackInfo---\n%s", file, line, formattedMsg, stack)
}

func CurrFuncName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	functionName := runtime.FuncForPC(pc).Name()
	ss := strings.Split(functionName, ".")
	if len(ss) > 2 {
		return strings.Join(ss[len(ss)-2:], ".")
	}
	return ss[len(ss)-1]
}

func JoinStr(ss ...string) string {
	return strings.Join(ss, "")
}
