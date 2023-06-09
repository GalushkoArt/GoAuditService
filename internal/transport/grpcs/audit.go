package grpcs

import (
	"context"
	"github.com/galushkoart/go-audit-service/internal/service"
	"github.com/galushkoart/go-audit-service/pkg/model"
	audit "github.com/galushkoart/go-audit-service/pkg/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type auditHandler struct {
	service service.AuditService
	audit.UnimplementedAuditServiceServer
}

func newAuditHandler(service service.AuditService) *auditHandler {
	return &auditHandler{service: service}
}

func (h *auditHandler) Log(ctx context.Context, request *audit.LogRequest) (*audit.Response, error) {
	err := h.service.Insert(ctx, model.LogRequestToItem(request))
	if err == nil {
		return &audit.Response{Answer: audit.Response_SUCCESS}, nil
	} else {
		log.Error().Str("from", "auditHandler").Interface("request", request).Msg("Failed to insert log!")
		return &audit.Response{Answer: audit.Response_ERROR}, status.Error(codes.Internal, err.Error())
	}
}
