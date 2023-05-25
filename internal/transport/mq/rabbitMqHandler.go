package mq

import (
	"context"
	"fmt"
	"github.com/GalushkoArt/GoAuditService/internal/config"
	"github.com/GalushkoArt/GoAuditService/internal/service"
	"github.com/GalushkoArt/GoAuditService/internal/utils"
	"github.com/GalushkoArt/GoAuditService/pkg/model"
	audit "github.com/GalushkoArt/GoAuditService/pkg/proto"
	"github.com/golang/protobuf/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"sync"
)

type Consumer struct {
	service   service.AuditService
	user      string
	password  string
	host      string
	port      int
	queueName string
}

var mqLog zerolog.Logger

func NewMqConsumer(service service.AuditService, conf config.MqConf) *Consumer {
	mqLog = log.With().Str("from", "mqConsumer").Logger()
	return &Consumer{
		service:   service,
		user:      conf.User,
		password:  conf.Password,
		host:      conf.Host,
		port:      conf.Port,
		queueName: conf.QueueName,
	}
}

func (s *Consumer) StartMqConsumer(enabled bool) func() {
	if !enabled {
		mqLog.Info().Msg("MQ consumer disabled")
		return func() {}
	}
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/", s.user, s.password, s.host, s.port))
	utils.PanicOnError(err, "Failed to connect to RabbitMQ")

	ch, err := conn.Channel()
	utils.PanicOnError(err, "Failed to open a channel")

	q, err := ch.QueueDeclare(
		s.queueName, // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	utils.PanicOnError(err, "Failed to declare a queue")
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	utils.PanicOnError(err, "Failed to register a consumer")

	ctx := context.Background()
	wg := &sync.WaitGroup{}

	go func() {
		for msg := range msgs {
			wg.Add(1)
			auditRequest := &audit.LogRequest{}
			err = proto.Unmarshal(msg.Body, auditRequest)
			if err != nil {
				mqLog.Error().Err(err).Msg("Fail unmarshal")
				utils.PanicOnError(msg.Nack(false, false))
				wg.Done()
				continue
			}
			mqLog.Debug().Msgf("Received msg: %v", auditRequest)
			err = s.service.Insert(ctx, model.LogRequestToItem(auditRequest))
			if err != nil {
				mqLog.Error().Err(err).Msgf("Fail to insert log")
				utils.PanicOnError(msg.Nack(false, false))
				wg.Done()
				continue
			}
			err = msg.Ack(false)
			if err != nil {
				mqLog.Error().Err(err).Msg("Fail to ack")
			}
			wg.Done()
		}
	}()

	mqLog.Info().Msg("MQ consumer started successfully")
	return func() {
		utils.PanicOnError(ch.Close())
		utils.PanicOnError(conn.Close())
		wg.Wait()
	}
}
