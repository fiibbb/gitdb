package handler

import (
	"context"
	"github.com/fiibbb/gitdb/config"
	"github.com/fiibbb/gitdb/consts"
	"github.com/fiibbb/gitdb/gitpb"
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
		grpc.MaxSendMsgSize(consts.MaxGRPCMessageSize),
		grpc.MaxRecvMsgSize(consts.MaxGRPCMessageSize),
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

func (s *GRPCService) WriteCommit(ctx context.Context, req *gitpb.WriteCommitRequest) (*gitpb.WriteCommitResponse, error) {
	resp, err := s.gitHandler.WriteCommit(ctx, req.Repo, req.Ref, req.Upserts, req.Deletes, req.Msg)
	if err != nil {
		return nil, err
	}
	return &gitpb.WriteCommitResponse{Commit: resp}, consts.ErrNYI
}

func (s *GRPCService) GetObject(ctx context.Context, req *gitpb.GetObjectRequest) (*gitpb.GetObjectResponse, error) {
	resp, err := s.gitHandler.GetObject(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &gitpb.GetObjectResponse{Object: resp}, nil
}
