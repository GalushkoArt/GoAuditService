package service

import (
	"context"
	"github.com/galushkoart/go-audit-service/pkg/model"
)

type AuditService interface {
	Insert(ctx context.Context, item *model.LogItem) error
}
