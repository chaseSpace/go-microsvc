package utilcommon

import (
	"fmt"
	"log"
	"runtime"
)

func TraceStack(skip ...int) string {
	errStack := ""
	pc := make([]uintptr, 6) // 最多保留

	_skip := 5
	if len(skip) > 0 {
		_skip = skip[0]
	}
	n := runtime.Callers(_skip, pc)
	frames := runtime.CallersFrames(pc[:n])

	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		errStack += fmt.Sprintf("%s:%d Func: %s\n-\n", frame.File, frame.Line, frame.Function)
	}
	return errStack
}

func TryCatch(f func(err interface{}, stack string)) {
	if e := recover(); e != nil {
		stack := TraceStack(4)
		log.Println("panic ------------\n", stack)
		f(e, stack)
	}
}
