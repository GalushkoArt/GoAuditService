package tests

import (
	"context"
	"flag"
	"fmt"
	"github.com/galushkoart/go-audit-service/internal/app"
	"github.com/galushkoart/go-audit-service/internal/config"
	"github.com/galushkoart/go-audit-service/internal/repository"
	"github.com/galushkoart/go-audit-service/internal/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

type IntegrationTests struct {
	AuditRepository repository.AuditRepository
	ctx             context.Context
	mongoContainer  *utils.Container
	mqContainer     *utils.Container
	grpcShutdown    func()
	mqShutdown      func()
	dbClient        *mongo.Client
	database        *mongo.Database
}

var it = &IntegrationTests{}
var exitMain = make(chan bool)
var finishedMain = make(chan bool)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Short() {
		testing.Short()
		it.SetupSuite()
		rc := m.Run()
		exitMain <- true
		<-finishedMain
		it.TearDownSuite()
		os.Exit(rc)
	} else {
		os.Exit(m.Run())
	}
}

func (it *IntegrationTests) SetupSuite() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	utils.PanicOnError(config.InitTest())
	var err error
	it.ctx = context.Background()
	waitContainers := sync.WaitGroup{}
	waitContainers.Add(2)
	go func() {
		it.mongoContainer, err = utils.PrepareContainer(it.ctx, utils.MongoDB, config.TestConf.Mongo.Version)
		utils.PanicOnError(err, "Failed to prepare MongoDB container!")
		waitContainers.Done()
	}()
	go func() {
		it.mqContainer, err = utils.PrepareContainer(it.ctx, config.TestConf.MQ.Image, config.TestConf.MQ.Version)
		utils.PanicOnError(err, "Failed to prepare MQ container!")
		waitContainers.Done()
	}()
	waitContainers.Wait()
	utils.PanicOnError(os.Setenv("MONGO_URI", fmt.Sprintf("mongodb://%s:%d", it.mongoContainer.HostIP, it.mongoContainer.Port)))
	utils.PanicOnError(os.Setenv("MQ_HOST", it.mqContainer.HostIP))
	utils.PanicOnError(os.Setenv("MQ_PORT", strconv.Itoa(it.mqContainer.Port)))
	started := make(chan bool)
	go app.Run(exitMain, started, finishedMain)
	<-started
	it.dbClient, err = mongo.Connect(it.ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%d", it.mongoContainer.HostIP, it.mongoContainer.Port)).SetTimeout(10*time.Second))
	utils.PanicOnError(err, "Failed to connect to MongoDB")
	it.database = it.dbClient.Database("audit")
	it.AuditRepository = repository.NewAuditRepository(it.database)
}

func (it *IntegrationTests) TearDownSuite() {
	done := make(chan bool)
	finishWG := sync.WaitGroup{}
	finishWG.Add(2)
	go func() {
		utils.PanicOnError(it.dbClient.Disconnect(it.ctx), "Failed to disconnect from MongoDB!")
		utils.PanicOnError(it.mongoContainer.Terminate(it.ctx), "Failed to terminate MongoDB container!")
		finishWG.Done()
	}()
	go func() {
		err := it.mqContainer.Terminate(it.ctx)
		log.Error().Err(err).Msg("Failed to terminate MQ container!")
		finishWG.Done()
	}()
	go func() {
		finishWG.Wait()
		done <- true
	}()
	<-done
}
