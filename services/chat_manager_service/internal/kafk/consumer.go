package kafk

import (
	"chat_manager_service/internal/models"
	"chat_manager_service/internal/service"
	"context"
	"encoding/json"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

type Consumer struct {
	cons       *kafka.Consumer
	service    *service.Service
	consLogger *zap.Logger
	ctx        context.Context
}

func NewConsumer(ctx context.Context, service *service.Service, consLogger *zap.Logger) *Consumer {
	cons, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  "localhost:",
		"group.id":           "chat_manager",
		"enable.auto.commit": false,
		"auto.offset.reset":  true,
	})
	if err != nil {
		panic(err)
	}
	if err := cons.Subscribe("new_user", nil); err != nil {
		panic(err)
	}
	return &Consumer{
		ctx:        ctx,
		cons:       cons,
		service:    service,
		consLogger: consLogger,
	}
}

func (c *Consumer) AddNewUser(msg *kafka.Message) error {
	var userData models.AddNewUser
	if err := json.Unmarshal(msg.Value, &userData); err != nil {
		//log
		return err
	}

	if err := c.service.AddNewUser(c.ctx, userData.Id, userData.Email); err != nil {
		//log
		return err
	}
	return nil
}

func (c *Consumer) Consume() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			time.Sleep(500 * time.Millisecond)
			msg, err := c.cons.ReadMessage(-1)
			if err != nil {
				//log
				continue
			}

			switch *msg.TopicPartition.Topic {
			case "new_user":
				err = c.AddNewUser(msg)
			default:
			}
			if err != nil {
				//log
			}
			if _, err := c.cons.CommitMessage(msg); err != nil {
				//log
			}
		}
	}
}
