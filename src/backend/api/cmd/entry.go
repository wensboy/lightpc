package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
	"github.com/wensboy/lightpc/backend/bootstrap"
	domainConfig "github.com/wensboy/lightpc/backend/domain/config"
	domainContext "github.com/wensboy/lightpc/backend/domain/context"
	"github.com/wensboy/lightpc/backend/internal/cast"
)

// Execute 传入嵌入上下文执行程序后返回程序执行状态码.
func Execute(eCtx domainContext.EmbedContext) int {
	cfg := bootstrap.SetupConfig(eCtx)
	flagSource := cfg.Sources(domainConfig.SourceFlag)[0]
	v, ok := flagSource.Lookup("lightpc")
	if !ok {
		fmt.Fprintln(os.Stderr, "程序无法通过 cli 接口执行")
		return 1
	}
	rootCmd, ok := cast.To[*cli.Command](v)
	if !ok {
		fmt.Fprintln(os.Stderr, "根指令解析失败")
		return 1
	}
	appCtx := domainContext.AppContext{
		Context:       context.Background(),
		EmbedContext:  eCtx,
		ConfigContext: cfg,
		ErrorContext:  bootstrap.SetupErrors(cfg),
	}
	setupCommand(rootCmd, appCtx)
	if err := rootCmd.Run(appCtx, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}

func setupCommand(root *cli.Command, appCtx domainContext.AppContext) {
	// lightpc
	root.Action = func(ctx context.Context, c *cli.Command) error {
		return bootstrap.Boot(domainContext.BootContext{
			Mode:       uint8(c.Int("mode")),
			AppContext: appCtx,
		})
	}
}
