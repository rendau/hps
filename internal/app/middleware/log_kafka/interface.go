package log_kafka

import (
	"context"

	"github.com/rendau/hps/internal/app/middleware/log_kafka/producer"
)

type filterI interface {
	Check(method, pathStr string) bool
}

type producerI interface {
	Send(ctx context.Context, items ...producer.Message) error
}
