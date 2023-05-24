package repository

import (
	"GoAuditService/pkg/model"
	"context"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuditRepository interface {
	Insert(ctx context.Context, item model.LogItem) error
}

type auditRepository struct {
	db *mongo.Database
}

func NewAuditRepository(db *mongo.Database) AuditRepository {
	return &auditRepository{
		db: db,
	}
}

func (r *auditRepository) Insert(ctx context.Context, item model.LogItem) error {
	log.Debug().Str("from", "auditRepository").Msgf("Inserting %v", item)
	_, err := r.db.Collection("logs").InsertOne(ctx, item)
	return err
}
