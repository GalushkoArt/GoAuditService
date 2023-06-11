package tests

import (
	"github.com/galushkoart/go-audit-service/pkg/model"
	audit "github.com/galushkoart/go-audit-service/pkg/proto"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"testing"
	"time"
)

func TestGRPCServer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	t.Parallel()
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := audit.NewAuditServiceClient(conn)
	require.NoError(t, err)
	log.Info().Msg("grpc client started")
	finishWg := sync.WaitGroup{}
	finishWg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			item := model.LogItem{Entity: model.ENTITY_USER, EntityID: uuid.New().String(), Action: model.ACTION_SIGN_IN, Timestamp: time.Now().UTC().Round(time.Millisecond)}
			response, err := client.Log(it.ctx, model.LogItemToLogRequest(&item))
			assert.NoError(t, err, "failed to send message")
			assert.Equal(t, response.Answer, audit.Response_SUCCESS, "response should be success")
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
