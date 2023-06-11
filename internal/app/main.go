package app

import (
	"context"
	"github.com/galushkoart/go-audit-service/internal/config"
	"github.com/galushkoart/go-audit-service/internal/logs"
	"github.com/galushkoart/go-audit-service/internal/repository"
	"github.com/galushkoart/go-audit-service/internal/transport/grpcs"
	"github.com/galushkoart/go-audit-service/internal/transport/mq"
	"github.com/galushkoart/go-audit-service/internal/utils"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

func Run(exit <-chan bool, started chan<- bool, finished chan<- bool) {
	config.Init()
	logs.Init(config.Conf.Logs.Level, config.Conf.Logs.Path)
	log.Debug().Msg("Configs and logs initialized")
	opts := options.Client()
	mongoConf := config.Conf.Mongo
	if mongoConf.Username != "" && mongoConf.Password != "" {
		opts.SetAuth(options.Credential{
			Username: mongoConf.Username,
			Password: mongoConf.Password,
		})
	}
	opts.ApplyURI(mongoConf.URI)
	ctx := context.Background()
	dbClient, err := mongo.Connect(ctx, opts)
	utils.PanicOnError(err, "Failed to connect to MongoDB")
	utils.PanicOnError(dbClient.Ping(ctx, nil), "Failed to ping MongoDB")
	db := dbClient.Database(mongoConf.Database)
	log.Info().Msgf("Connected to %s mongo database successfully!", mongoConf.Database)

	auditRepository := repository.NewAuditRepository(db)
	grpcShutdown := grpcs.StartGRPC(config.Conf.GRPC.Enabled, auditRepository, config.Conf.GRPC.Port)
	mqConsumer := mq.NewMqConsumer(auditRepository, config.Conf.MQ)
	mqShutdown := mqConsumer.StartMqConsumer(config.Conf.MQ.Enabled)

	log.Info().Msg("Server was started")
	started <- true

	<-exit

	done := make(chan bool)
	finishWG := sync.WaitGroup{}
	finishWG.Add(3)
	go func() {
		grpcShutdown()
		finishWG.Done()
	}()
	go func() {
		mqShutdown()
		finishWG.Done()
	}()
	go func() {
		utils.PanicOnError(dbClient.Disconnect(ctx), "Failed to disconnect from MongoDB")
		finishWG.Done()
	}()
	go func() {
		finishWG.Wait()
		done <- true
	}()
	select {
	case <-time.After(30 * time.Second):
		log.Error().Msg("Failed to shutdown in 30 seconds")
		finished <- false
	case <-done:
		log.Info().Msg("Shutdown successfully")
		finished <- true
	}
}
