package handler

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"microsvc/pkg/xlog"
	"microsvc/util"
	"reflect"
	"time"
)

func MustInit(cmd *cobra.Command) {
	// 这里的逻辑很简单，利用反射来调用 ctrl{} 绑定的外部方法

	tp := reflect.TypeOf(ctrl{})
	val := reflect.ValueOf(ctrl{})

	for i := 0; i < tp.NumMethod(); i++ {
		funcName := tp.Method(i).Name
		i2 := i

		checkMethodSignature(tp.Method(i))

		cmd.AddCommand(&cobra.Command{
			Use: funcName,
			Run: func(cmd *cobra.Command, args []string) {
				st := time.Now()
				var err error
				xlog.Info("Start Execute...", zap.String("func", funcName), zap.Any("args", args))
				defer func() {
					zaps := []zap.Field{
						zap.Error(err),
						zap.String("func", funcName),
						zap.String("cost", time.Since(st).String()),
					}
					if err != nil {
						panic(err)
					}
					xlog.Info("Execute Success", zaps...)
				}()
				var vals []reflect.Value
				for _, v := range []any{context.TODO(), cmd, args} {
					vals = append(vals, reflect.ValueOf(v))
				}

				// 第一个出参只能是一个error
				errs := val.Method(i2).Call(vals)
				if len(errs) > 0 && !errs[0].IsNil() {
					err = errs[0].Interface().(error)
				}
			},
		})
	}

	util.AssertNilErr(cmd.Execute())
}

// Verify the method signature by comparing the method's type using type literals
func checkMethodSignature(me reflect.Method) {
	// 必须满足的Method出参入参规则
	var __methodNumIn = 4
	var __methodInParams = []string{"context.Context", "*cobra.Command", "[]string"}

	var __methodNumOut = 1
	var __methodOutParams = []string{"error"}

	if num := me.Type.NumIn(); num != __methodNumIn {
		panic(fmt.Sprintf("method [%s] signature error: wrong number of in [%d]", me.Name, num))
	}
	for j := 1; j < __methodNumIn; j++ {
		str := me.Type.In(j).String()
		if str != __methodInParams[j-1] {
			panic(fmt.Sprintf("method [%s] signature error: wrong type of in [%d:%s]", me.Name, j, str))
		}
	}

	if num := me.Type.NumOut(); num != __methodNumOut {
		panic(fmt.Sprintf("method [%s] signature error: wrong number of out [%d]", me.Name, num))
	}
	for j := 0; j < __methodNumOut; j++ {
		str := me.Type.Out(j).String()
		if str != __methodOutParams[j] {
			panic(fmt.Sprintf("method [%s] signature error: wrong type of out [%d:%s]", me.Name, j, str))
		}
	}
}
