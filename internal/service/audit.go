package service

import (
	"context"
	"github.com/GalushkoArt/GoAuditService/pkg/model"
)

type AuditService interface {
	Insert(ctx context.Context, item model.LogItem) error
}
