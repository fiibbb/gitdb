package handler

import (
	"github.com/fiibbb/gitdb/config"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"go.uber.org/zap"
	"net/http"
)

type HTTPService struct {
	cfg        *config.AppConfig
	gatewayMux *gwruntime.ServeMux
	server     *http.Server
	logger     *zap.Logger
}

func NewHTTPService(cfg *config.AppConfig, logger *zap.Logger) (*HTTPService, error) {
	gatewayMux := gwruntime.NewServeMux()
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: gatewayMux,
	}
	return &HTTPService{
		cfg:        cfg,
		gatewayMux: gatewayMux,
		server:     server,
		logger:     logger,
	}, nil
}
