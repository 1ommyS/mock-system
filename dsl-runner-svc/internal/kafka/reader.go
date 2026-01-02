package kafka

import "github.com/segmentio/kafka-go"

func NewReader(brokers []string, topic, group string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     group,
		StartOffset: kafka.FirstOffset,
		MinBytes:    1e4,
		MaxBytes:    1e6,
	})
}
