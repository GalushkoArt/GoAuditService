package main

import (
	"GoAuditService/internal/config"
	"GoAuditService/internal/logs"
	"GoAuditService/internal/repository"
	"GoAuditService/internal/transport/gRPC"
	"GoAuditService/internal/transport/mq"
	"GoAuditService/internal/utils"
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	config.Init()
	logs.Init(config.Conf.Logs.Level, config.Conf.Logs.Path)
	log.Debug().Msg("Configs and logs initialized")
	opts := options.Client()
	mongoConf := config.Conf.Mongo
	opts.SetAuth(options.Credential{
		Username: mongoConf.Username,
		Password: mongoConf.Password,
	})
	opts.ApplyURI(mongoConf.URI)
	dbClient, err := mongo.Connect(context.Background(), opts)
	utils.PanicOnError(err, "Failed to connect to MongoDB")
	utils.PanicOnError(dbClient.Ping(context.Background(), nil), "Failed to ping MongoDB")
	db := dbClient.Database(mongoConf.Database)
	log.Info().Msgf("Connected to %s mongo database successfully!", mongoConf.Database)

	auditRepository := repository.NewAuditRepository(db)
	grpcShutdown := gRPC.StartGRPC(config.Conf.GRPC.Enabled, auditRepository, config.Conf.GRPC.Port)
	mqConsumer := mq.NewMqConsumer(auditRepository, config.Conf.MQ)
	mqShutdown := mqConsumer.StartMqConsumer(config.Conf.MQ.Enabled)

	log.Info().Msg("Server was started")

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, os.Kill)
	fmt.Println(<-exit)

	done := make(chan bool)
	finishWG := sync.WaitGroup{}
	finishWG.Add(2)
	go func() {
		grpcShutdown()
		finishWG.Done()
	}()
	go func() {
		mqShutdown()
		finishWG.Done()
	}()
	go func() {
		finishWG.Wait()
		done <- true
	}()
	select {
	case <-time.After(30 * time.Second):
		log.Error().Msg("Failed to shutdown in 30 seconds")
		os.Exit(1)
	case <-done:
		log.Info().Msg("Shutdown successfully")
	}
}
