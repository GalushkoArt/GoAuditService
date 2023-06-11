package repository

import (
	"context"
	"github.com/galushkoart/go-audit-service/pkg/model"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuditRepository interface {
	Insert(ctx context.Context, item *model.LogItem) error
}

type auditRepository struct {
	db *mongo.Database
}

func NewAuditRepository(db *mongo.Database) AuditRepository {
	return &auditRepository{
		db: db,
	}
}

func (r *auditRepository) Insert(ctx context.Context, item *model.LogItem) error {
	res, err := r.db.Collection("logs").InsertOne(ctx, &item)
	log.Debug().Str("from", "auditRepository").Interface("result", res).Msg("Insert result")
	return err
}
