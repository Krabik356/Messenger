package kafk

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Producer struct {
	prod *kafka.Producer
}

func NewProducer() *Producer {
	prod, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":  "localhost:",
		"acks":               "all",
		"enable.idempotence": true,
	})
	if err != nil {
		panic(err)
	}
	return &Producer{
		prod: prod,
	}
}
