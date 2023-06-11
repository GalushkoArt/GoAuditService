package grpcs

import (
	"context"
	audit "github.com/galushkoart/go-audit-service/pkg/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"time"
)

func requestLogger(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	request := req.(*audit.LogRequest)
	response := resp.(*audit.Response)
	log.Info().
		Err(err).
		Str("status", response.Answer.String()).
		Str("request-id", request.RequestId).
		Str("latency", time.Since(start).String()).
		Msg("request")

	return resp, err
}
