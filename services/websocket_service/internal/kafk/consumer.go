package kafk

import (
	"context"
	"encoding/json"
	"messenger/services/websocket_service/internal/service"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

type Consumer struct {
	cons       *kafka.Consumer
	consLogger *zap.Logger
	service    *service.Service
	ctx        context.Context
}

func NewConsumer(ctx context.Context) *Consumer {
	cons, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  "localhost:9091",
		"group.id":           "websocket_service",
		"enable.auto.commit": false,
		"auto.offset.reset":  "earliest",
	})
	if err != nil {
		panic(err)
	}
	if err := cons.Subscribe("new_chat", nil); err != nil {
		panic(err)
	}
	if err := cons.Subscribe("new_message", nil); err != nil {
		panic(err)
	}
	return &Consumer{
		cons: cons,
		ctx:  ctx,
	}
}

func (c *Consumer) logConsumer(key, status, err string) {
	if err != "" {
		c.consLogger.Error("log",
			zap.String("key", key),
			zap.String("status", status),
			zap.String("error", err))
		return
	}
	c.consLogger.Info("log",
		zap.String("key", key),
		zap.String("status", status),
		zap.String("error", "nil"))
}

func (c *Consumer) Consume() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msg, err := c.cons.ReadMessage(-1)
			if err != nil {
				c.logConsumer("-", "error", err.Error())
				continue
			}

			if *msg.TopicPartition.Topic == "new_chat" {
				var userIds []int

				chatId, err := strconv.Atoi(string(msg.Key))
				if err != nil {
					c.logConsumer("-", "error", "error with unmarshalling chat id")
					continue
				}
				if err := json.Unmarshal(msg.Value, &userIds); err != nil {
					c.logConsumer("-", "error", "error with unmarshalling user ids")
					continue
				}

				if err := c.service.NewChat(c.ctx, chatId, userIds...); err != nil {
					c.logConsumer(strconv.Itoa(chatId), "error", err.Error())
					continue
				}
			} else if *msg.TopicPartition.Topic == "new_message" {
				//
			}
		}
	}
}
