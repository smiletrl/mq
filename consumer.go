package mq

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
)

var (
	consumerMu sync.RWMutex

	// new consumer should register
	consumers = make(map[string]Consumer)
)

type Consumer interface {
	// consumer name. one recommended format `{service_name}:{internal_name}:{event_name}`
	Name() string

	// which event this consumer subscribes to
	Event() Event

	// consumer can decide whether it wants to consume its message in delayed time
	Delay() time.Duration

	// consume one message
	Consume(ctx context.Context, tx pgx.Tx, msg *MQMessage) error
}

func RegisterConsumer(consumer Consumer) {
	consumerMu.Lock()
	defer consumerMu.Unlock()
	if _, dup := consumers[consumer.Name()]; dup {
		panic("consumer register called twice for " + consumer.Name())
	}
	consumers[consumer.Name()] = consumer
}
