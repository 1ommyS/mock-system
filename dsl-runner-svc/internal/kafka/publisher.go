package kafka

import (
	"context"
	"encoding/json"

	"dsl-runner-svc/internal/domain"

	"github.com/segmentio/kafka-go"
)

type ResultPublisher struct {
	Writer *kafka.Writer
	Topic  string
}

func NewResultPublisher(brokers []string, topic string) *ResultPublisher {
	return &ResultPublisher{
		Writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafka.Hash{},
			AllowAutoTopicCreation: true,
		},
		Topic: topic,
	}
}

func (p *ResultPublisher) Publish(ctx context.Context, msg domain.ResultMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	kmsg := kafka.Message{
		Key:   []byte(msg.JobKey),
		Value: data,
	}
	return p.Writer.WriteMessages(ctx, kmsg)
}
