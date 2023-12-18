package mq

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

type Provider interface {
	// send message with pgx pool tx
	SendMessage(ctx context.Context, tx pgx.Tx, message Message) error
}

func NewProvider(tenantPool *pgxpool.Pool) Provider {
	return provider{
		pool: tenantPool,
	}
}

type provider struct {
	// group is used to create messages. One event/tag(e.g, order_created) will insert every consumer group's own message
	consumers map[Event][]Consumer
	pool      *pgxpool.Pool
}

// Lazy loading. innner message group
// For one event, such as `order_created`, service notify and order subscribe to this event.
// When this event `order_created` happens, we will create two messages. One message is for service notify's consumer,
// while the other one is for service order's consumer.
// In above case, event `order_created` will have a group of consumers: notify consumer and order consumer.
// So here we have a map data structure `map[Event][]Consumer`,  key is the event, and value is a slice of consumers.
func (p provider) innerConsumers() map[Event][]Consumer {
	if p.consumers != nil {
		return p.consumers
	}

	innerConsumers := make(map[Event][]Consumer)
	for _, consumer := range consumers {
		if _, ok := innerConsumers[consumer.Event()]; !ok {
			innerConsumers[consumer.Event()] = []Consumer{consumer}
		} else {
			innerConsumers[consumer.Event()] = append(innerConsumers[consumer.Event()], consumer)
		}
	}
	p.consumers = innerConsumers
	return p.consumers
}

func (p provider) SendMessage(ctx context.Context, tx pgx.Tx, message Message) error {
	event := message.Event()
	consumerGroups, ok := p.innerConsumers()[event]
	if !ok {
		return errors.Errorf("mq event: %s does not have consumer groups", event.String())
	}
	query := `insert into queues(consumer_name, message, check_at) values`

	var (
		index int
		args  []interface{}
	)
	createdAt := time.Now()

	for i, consumer := range consumerGroups {
		val := fmt.Sprintf("($%d, $%d, $%d)", index+1, index+2, index+3)
		index = index + 3
		if i < len(consumerGroups)-1 {
			query = query + val + `,`
		} else {
			query = query + val
		}
		checkAt := createdAt.Add(consumer.Delay())
		args = append(args, consumer.Name(), message, checkAt)
	}
	if _, err := tx.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("error inserting message queue: %w", err)
	}
	return nil
}
