package main

import (
	"github.com/fiibbb/gitdb/config"
	"github.com/fiibbb/gitdb/handler"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(
			zap.NewDevelopment,
			config.NewAppConfig,
			handler.NewGitHandler,
			handler.NewGRPCService,
			handler.NewHTTPService,
		),
		fx.Invoke(
			handler.Serve,
		),
	).Run()
}
