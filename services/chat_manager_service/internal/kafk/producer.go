package kafk

import (
	"chat_manager_service/internal/models"
	"chat_manager_service/internal/service"
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

type Producer struct {
	prod       *kafka.Producer
	service    *service.Service
	ctx        context.Context
	stopCtx    context.CancelFunc
	prodLogger *zap.Logger
}

func NewProducer(service *service.Service) *Producer {
	prod, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":  "localhost:9091",
		"acks":               "all",
		"enable.idempotence": true,
	})
	if err != nil {
		panic(err)
	}
	ctx, stop := context.WithCancel(context.Background())
	return &Producer{
		prod:    prod,
		service: service,
		ctx:     ctx,
		stopCtx: stop,
	}
}

func (p *Producer) Close() {
	p.prod.Flush(5000)
	p.prod.Close()
	p.stopCtx()
}

func (p *Producer) logProducer(key, status, err string) {
	if err != "" {
		p.prodLogger.Error("log",
			zap.String("key", key),
			zap.String("status", status),
			zap.String("error", err))
		return
	}
	p.prodLogger.Info("log",
		zap.String("key", key),
		zap.String("status", status),
		zap.String("error", "nil"))
}

func (p *Producer) ProduceNewMessage(rawMsg models.Message) error {
	msg, err := json.Marshal(rawMsg)
	if err != nil {
		return models.ServersError
	}

	if err := p.prod.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     new("messages"),
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(strconv.Itoa(rawMsg.Id)),
		Value: msg,
	}, nil); err != nil {
		return models.ServersError
	}

	return nil
}

func (p *Producer) Produce() {
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			time.Sleep(500 * time.Millisecond)
			messages, err := p.service.GetFromOutbox(p.ctx)
			if err != nil {
				p.logProducer("-", "error", err.Error())
			}

			for _, msg := range messages {
				if err := p.ProduceNewMessage(msg); err != nil {
					p.logProducer(strconv.Itoa(msg.Id), "error", err.Error())
				}
			}

		}
	}
}

func (p *Producer) EventListener() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case e := <-p.prod.Events():
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					p.logProducer(string(ev.Key), "error", "error with sending")
					continue
				}
				id, err := strconv.Atoi(string(ev.Key))
				if err != nil {
					p.logProducer(string(ev.Key), "error", "error with converting id to int")
					continue
				}
				if err := p.service.CommitMessage(p.ctx, id); err != nil {
					p.logProducer(string(ev.Key), "error", "error with commiting")
					continue
				}
				p.logProducer(string(ev.Key), "", "success")
			case kafka.Error:
				p.logProducer("-", "error", ev.Error())
				continue
			}
		}
	}
}
