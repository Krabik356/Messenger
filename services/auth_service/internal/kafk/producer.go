package kafk

import (
	"Messenger/internal/models"
	"Messenger/internal/service"
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Producer struct {
	prod    *kafka.Producer
	service *service.Service
	ctx     context.Context
	stopCtx context.CancelFunc
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
					//log
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
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					//log
					continue
				}
				id, err := strconv.Atoi(string(ev.Key))
				if err != nil {
					//log
					continue
				}
				if err := p.service.CommitOutboxByUserId(p.ctx, id); err != nil {
					//log
					continue
				}
			case kafka.Error:
				//log
				continue
			}
		}
	}
}
