package tests

import (
	"github.com/galushkoart/go-audit-service/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"
)

func TestAuditRepositoryInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	t.Parallel()
	inserted := model.LogItem{Action: model.ACTION_SIGN_UP, EntityID: uuid.New().String(), Entity: model.ENTITY_USER, Timestamp: time.Now().UTC().Round(time.Millisecond)}
	err := it.AuditRepository.Insert(it.ctx, &inserted)
	require.NoError(t, err, "No errors on successful insert to db")
	found := &model.LogItem{}
	err = it.database.Collection("logs").FindOne(it.ctx, bson.M{"entity_id": inserted.EntityID}).Decode(found)
	assert.NoError(t, err, "Item should be found and decoded")
	assert.Equal(t, inserted, *found)
}
