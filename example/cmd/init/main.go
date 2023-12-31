package main

import (
	"context"
	"log"

	"github.com/smiletrl/mq/example/pkg/postgres"
)

func main() {
	// pgx db pool
	pool, err := postgres.NewPool()
	if err != nil {
		panic("pgx pool instance error: " + err.Error())
	}

	ctx := context.Background()

	if _, err := pool.Exec(ctx, initdb); err != nil {
		log.Printf("error init db: %v", err)
	} else {
		log.Println("init db successfully")
	}
}

const initdb = `
    CREATE TABLE IF NOT EXISTS queues (
        id bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
        consumer_name text NOT NULL,
        message jsonb NOT NULL,
        retry int DEFAULT 0 NOT NULL,
        is_dead boolean DEFAULT false NOT NULL,
        failed_reason text,
        check_at timestamp NOT NULL,

        created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

    comment on column queues.consumer_name is 'which consumer should be used to consume this message';
    comment on column queues.message is 'one json string containing message detail, like {order_id: 12}';
    comment on column queues.retry is 'the number of times this message has been consumed';
    comment on column queues.is_dead is 'when this message is dead, it means this message has reached max retry times, yet still failed to be consumed';
    comment on column queues.failed_reason is 'log the failed message when consumer fails to consume this message';
    comment on column queues.check_at is 'when cron system should check this message and consume it';

    CREATE TABLE IF NOT EXISTS orders (
        id bigint GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY,
        product text DEFAULT '',
        status text check (status in ('pending_payment', 'canceled')),
        created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
    );

    comment on column orders.product is 'what product is ordered';
    comment on column orders.status is 'order status. For demo purpose, it only allows value pending_payment/canceled';
	`
