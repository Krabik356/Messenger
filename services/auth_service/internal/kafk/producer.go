package kafk

import (
	"Messenger/internal/models"
	"Messenger/internal/service"
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

type Producer struct {
	prod           *kafka.Producer
	service        *service.Service
	producerLogger *zap.Logger
	ctx            context.Context
	stopCtx        context.CancelFunc
}

func NewProducer(service *service.Service) *Producer {
	prod, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":  "localhost:",
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
		p.producerLogger.Error("log",
			zap.String("key", key),
			zap.String("status", status),
			zap.String("error", err))
		return
	}
	p.producerLogger.Info("log",
		zap.String("key", key),
		zap.String("status", status),
		zap.String("error", "nil"))
}

func (p *Producer) ProduceNewUser(id int, name, email string) error {
	data, err := json.Marshal(models.RegForKafka{
		Id:    id,
		Name:  name,
		Email: email,
	})
	if err != nil {
		return models.ServersError
	}

	return p.prod.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     new("new_user"),
			Partition: kafka.PartitionAny,
		},
		Value: data,
		Key:   []byte(strconv.Itoa(id)),
	}, nil)
}

func (p *Producer) Produce() {
	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			time.Sleep(500 * time.Millisecond)
			data, err := p.service.GetFromOutbox(p.ctx, 10)
			if err != nil {
				//log
				continue
			}
			for _, user := range data {
				if err := p.ProduceNewUser(user.Id, user.Name, user.Email); err != nil {
					p.logProducer(string(user.Id), "error_on_producing", err.Error())
					continue
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
			var stringId string
			switch ev := e.(type) {
			case *kafka.Message:
				stringId = string(ev.Key)
				if ev.TopicPartition.Error != nil {
					p.logProducer(stringId, "error_on_sending", ev.TopicPartition.Error.Error())
					continue
				}
				id, err := strconv.Atoi(stringId)
				if err != nil {
					p.logProducer(stringId, "error_on_id_parsing", err.Error())
					continue
				}
				if err := p.service.CommitOutboxByUserId(p.ctx, id); err != nil {
					p.logProducer(stringId, "error_on_commiting", err.Error())
					continue
				}
			case kafka.Error:
				p.logProducer("-", "error_on_connecting", ev.Error())
				continue
			}
			p.logProducer(stringId, "success", "")
		}
	}
}
