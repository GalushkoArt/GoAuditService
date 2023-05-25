package gRPC

import (
	"fmt"
	"github.com/GalushkoArt/GoAuditService/internal/service"
	"github.com/GalushkoArt/GoAuditService/internal/utils"
	audit "github.com/GalushkoArt/GoAuditService/pkg/proto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"time"
)

type Server struct {
	grpcSrv     *grpc.Server
	auditServer *auditHandler
}

var grpcLog zerolog.Logger

func StartGRPC(enabled bool, auditService service.AuditService, port int) func() {
	grpcLog = log.With().Str("from", "grpcServer").Logger()
	if enabled {
		return New(auditService).ListenAndServe(port)
	} else {
		grpcLog.Info().Msg("gRPC server disabled")
		return func() {}
	}
}

func New(service service.AuditService) *Server {
	kaep := keepalive.EnforcementPolicy{
		MinTime:             10 * time.Second,
		PermitWithoutStream: true,
	}
	kasp := keepalive.ServerParameters{
		Time:    5 * time.Minute,
		Timeout: 5 * time.Minute,
	}
	return &Server{
		grpcSrv: grpc.NewServer(
			grpc.KeepaliveEnforcementPolicy(kaep),
			grpc.KeepaliveParams(kasp),
			grpc.UnaryInterceptor(requestLogger),
		),
		auditServer: newAuditHandler(service),
	}
}

func (s *Server) ListenAndServe(port int) func() {
	addr := fmt.Sprintf(":%d", port)

	lis, err := net.Listen("tcp", addr)
	utils.PanicOnError(err, fmt.Sprintf("Failed to listen port :%d", port))
	audit.RegisterAuditServiceServer(s.grpcSrv, s.auditServer)

	go func() {
		if err := s.grpcSrv.Serve(lis); err != nil {
			grpcLog.Info().Err(err).Msg("gRPC server has stopped")
		}
	}()

	grpcLog.Info().Msgf("gRPC server started successfully on port %s", addr)
	return s.grpcSrv.GracefulStop
}
