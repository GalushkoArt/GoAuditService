package service

import (
	"GoAuditService/pkg/model"
	"context"
)

type AuditService interface {
	Insert(ctx context.Context, item model.LogItem) error
}
