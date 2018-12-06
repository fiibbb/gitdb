package handler

import (
	"context"
	"github.com/fiibbb/gitdb/config"
	"github.com/fiibbb/gitdb/consts"
	"github.com/fiibbb/gitdb/gitpb"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

func Serve(cfg *config.AppConfig, g *GRPCService, h *HTTPService, lifecycle fx.Lifecycle) error {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			gitpb.RegisterGitServer(g.grpcServer, g)
			lis, err := net.Listen("tcp", cfg.GRPCAddr)
			if err != nil {
				return errors.WithStack(err)
			}
			if err := gitpb.RegisterGitHandlerFromEndpoint(context.Background(), h.gatewayMux, cfg.GRPCAddr, []grpc.DialOption{grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(consts.MaxGRPCMessageSize))}); err != nil {
				return errors.WithStack(err)
			}
			go g.grpcServer.Serve(lis)
			go h.server.ListenAndServe()
			g.logger.Info("Serving GRPC traffic", zap.String("addr", cfg.GRPCAddr))
			h.logger.Info("Serving HTTP traffic", zap.String("addr", cfg.HTTPAddr))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
	return nil
}
