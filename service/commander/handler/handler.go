package handler

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
)

/* 注意：

- 所有外部方法都是首字母大写
- 在这里的方法可以像普通服务那样操作数据库，用于临时查询、插入数据等操作
- 强制method签名：func(ctx context.Context, cmd *cobra.Command, args []string) error
*/

type ctrl struct {
}

func (ctrl) Test(ctx context.Context, cmd *cobra.Command, args []string) error {
	fmt.Println("run Test method...", args)
	return nil
}
