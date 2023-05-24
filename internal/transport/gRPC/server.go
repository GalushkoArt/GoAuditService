package gRPC

import (
	"GoAuditService/internal/service"
	"GoAuditService/internal/utils"
	audit "GoAuditService/pkg/proto"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"net"
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
	return &Server{
		grpcSrv: grpc.NewServer(
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
