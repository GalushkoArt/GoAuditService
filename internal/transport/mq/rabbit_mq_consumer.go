package mq

import (
	"context"
	"fmt"
	"github.com/galushkoart/go-audit-service/internal/config"
	"github.com/galushkoart/go-audit-service/internal/service"
	"github.com/galushkoart/go-audit-service/internal/utils"
	"github.com/galushkoart/go-audit-service/pkg/model"
	audit "github.com/galushkoart/go-audit-service/pkg/proto"
	"github.com/golang/protobuf/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Consumer struct {
	service     service.AuditService
	user        string
	password    string
	host        string
	port        int
	queueName   string
	concurrency int
}

var mqLog zerolog.Logger

func NewMqConsumer(service service.AuditService, conf config.MqConf) *Consumer {
	mqLog = log.With().Str("from", "mqConsumer").Logger()
	return &Consumer{
		service:     service,
		user:        conf.User,
		password:    conf.Password,
		host:        conf.Host,
		port:        conf.Port,
		queueName:   conf.QueueName,
		concurrency: conf.Concurrency,
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

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan bool)
	for i := 0; i < s.concurrency; i++ {
		go func() {
			defer func() {
				done <- true
			}()
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-msgs:
					if !ok {
						return
					}
					s.processMessage(msg)
				}
			}
		}()
	}

	mqLog.Info().Msg("MQ consumer started successfully")
	return func() {
		cancel()
		for i := 0; i < s.concurrency; i++ {
			<-done
		}
		close(done)
		utils.PanicOnError(ch.Close())
		utils.PanicOnError(conn.Close())
	}
}

func (s *Consumer) processMessage(msg amqp.Delivery) {
	auditRequest := &audit.LogRequest{}
	err := proto.Unmarshal(msg.Body, auditRequest)
	if err != nil {
		mqLog.Error().Err(err).Msg("Fail unmarshal")
		utils.PanicOnError(msg.Nack(false, false))
		return
	}
	mqLog.Debug().Msgf("Received msg: %v", auditRequest)
	err = s.service.Insert(context.Background(), model.LogRequestToItem(auditRequest))
	if err != nil {
		mqLog.Error().Err(err).Msgf("Fail to insert log")
		utils.PanicOnError(msg.Nack(false, false))
		return
	}
	err = msg.Ack(false)
	if err != nil {
		mqLog.Error().Err(err).Msg("Fail to ack")
	}
}
