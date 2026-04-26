package producer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type Service struct {
	writer *kafka.Writer
}

func New(host, topic string) *Service {
	if host == "" || topic == "" {
		return &Service{}
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(host),
		Topic:                  topic,
		RequiredAcks:           kafka.RequireOne,
		Async:                  true,
		AllowAutoTopicCreation: true,
	}

	slog.Info("Kafka writer created", "host", writer.Addr.String(), "topic", writer.Topic)

	return &Service{
		writer: writer,
	}
}

func (s *Service) Send(ctx context.Context, items ...Message) error {
	if s.writer == nil {
		return fmt.Errorf("writer is nil")
	}

	messages := make([]kafka.Message, 0, len(items))

	for _, item := range items {
		messages = append(messages, kafka.Message{
			Key:   []byte(item.Key),
			Value: item.Data,
		})
	}

	err := s.writer.WriteMessages(ctx, messages...)
	if err != nil {
		return fmt.Errorf("writer.WriteMessages: %w", err)
	}

	return nil
}

func (s *Service) Close() {
	if s.writer != nil {
		_ = s.writer.Close()
	}
}
