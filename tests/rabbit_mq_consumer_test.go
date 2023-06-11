package tests

import (
	"context"
	"fmt"
	"github.com/galushkoart/go-audit-service/pkg/model"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
	"testing"
	"time"
)

func TestRabbitMQConsumer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	conn, err := amqp.Dial(fmt.Sprintf("amqp://guest:guest@%s:%d/", it.mqContainer.HostIP, it.mqContainer.Port))
	require.NoError(t, err)
	channel, err := conn.Channel()
	require.NoError(t, err)
	log.Info().Msg("RabbitMQ publisher started")
	finishWg := sync.WaitGroup{}
	finishWg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			item := model.LogItem{Entity: model.ENTITY_USER, EntityID: uuid.New().String(), Action: model.ACTION_SIGN_IN, Timestamp: time.Now().UTC().Round(time.Millisecond)}
			data, err := proto.Marshal(model.LogItemToLogRequest(&item))
			assert.NoError(t, err, "failed to marshal request")
			err = channel.PublishWithContext(it.ctx, "", "audit", false, false, amqp.Publishing{
				ContentType: "text/plain",
				Body:        data,
			})
			assert.NoError(t, err, "failed to publish message")
			logItem := retrieveItem(it.database.Collection("logs"), item.EntityID)
			notNil := assert.NotNilf(t, logItem, "item %s should be found in db", item.EntityID)
			if notNil {
				assert.Equal(t, item, *logItem)
			}
			finishWg.Done()
		}()
	}
	finishWg.Wait()
}

func retrieveItem(collection *mongo.Collection, entityId string) *model.LogItem {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var result = &model.LogItem{}
	attempt := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case <-attempt.C:
			err := collection.FindOne(ctx, bson.M{"entity_id": entityId}).Decode(result)
			log.Trace().Err(err).Msgf("Attempt to find item %s in db", entityId)
			if err == nil {
				return result
			}
		case <-ctx.Done():
			attempt.Stop()
			return nil
		}
	}
}
