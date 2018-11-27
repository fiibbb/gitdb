package handler

import (
	"context"
	"github.com/fiibbb/gitdb/config"
	"github.com/fiibbb/gitdb/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GRPCService struct {
	cfg        *config.AppConfig
	grpcServer *grpc.Server
	gitHandler *GitHandler
	logger     *zap.Logger
}

func NewGRPCService(cfg *config.AppConfig, gitHandler *GitHandler, logger *zap.Logger) (*GRPCService, error) {
	grpcServer := grpc.NewServer(
		grpc.MaxSendMsgSize(MaxGRPCMessageSize),
		grpc.MaxRecvMsgSize(MaxGRPCMessageSize),
		grpc.UnaryInterceptor(grpcInterceptor))
	return &GRPCService{
		cfg:        cfg,
		grpcServer: grpcServer,
		gitHandler: gitHandler,
		logger:     logger,
	}, nil
}

func grpcInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		info.Server.(*GRPCService).logger.Info("error", zap.Error(err), zap.Any("request", req), zap.Any("response", resp))
	}
	return resp, err
}

func (s *GRPCService) Health(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	err := s.gitHandler.Health(ctx)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *GRPCService) GetObject(ctx context.Context, req *gitdbpb.GetObjectRequest) (*gitdbpb.GetObjectResponse, error) {
	resp, err := s.gitHandler.GetObject(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &gitdbpb.GetObjectResponse{Object: resp}, nil
}
