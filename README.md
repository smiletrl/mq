# MQ

One Message Queue system based on Postgres table.

For projects with smaller traffic volume, it's usually not cost wise to use [kafka](https://kafka.apache.org/),
rocketmq(https://rocketmq.apache.org/) or similar mq system. But projects still need [Eventual consistency](https://en.wikipedia.org/wiki/Eventual_consistency) to enable distributed system, and decouple services.

Here's the solution to make use of postgres table.

It makes use of postgres [transaction save point](https://www.postgresql.org/docs/current/sql-savepoint.html)
to have two save points within message consuming process. Due to the nature of RDS transaction, this system is
easier to be implemented with postgres tx. [pgx](https://github.com/jackc/pgx) is used to in this system.

It's up to project to use appropriate postgres library.

# Design

This system is based on below table `queues`:

```
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
```

Table `queues` will hold messages to be consumed. There will be a cron running periodically to
scan this table and consume messages with related consumer.

This queue supports retry, delay, dead queue. If required, a queue dashboard can be created to demonstrate current list
queue messages for system administrator to review queue health.

# Scenario example

In commerce app, user puts one order but not paid yet. There are two things to be done after order is created:

- Order service. Cancel this order if order is not paid for 5 minutes or other time duration.
- Notify servvice. Send one email notification out for this order.

When this `event` (one new order is created) happens, we will insert two messages into table `queues`. Each message will
have its own consumer.

More details can be found at dir `example`. This example needs docker env to enable postgres db.

This example uses [echo](https://echo.labstack.com/) to create one demo rest API server.

Steps to see real effect

- cd example
- make db-start
- make app-start
- run `curl localhost:1325/api/order/product -X POST -d '{product: "closing"}'`

If all goes well, golang echo server console will output something like this

```
2023/12/19 15:41:34 notification consumer is processed successfully!
2023/12/19 15:41:34 order consumer is processed successfully!
```

Both postgres and go code uses UTC timezone. If your local Go env's timezone is UTC > UTC+0 such as UTC+8, you may observe
above result. Otherwise it depends on your timezone.

# Development

- How to produce event `order_created` messages

see `example/service.order/service.go`

```
	msg := mq.NewOrderMessage(mq.EventOrderCreated).
		WithOrderID(orderID).
		Encode(ctx)
	if err := s.mq.SendMessage(ctx, tx, msg); err != nil {
		return fmt.Errorf("error sending mq message: %w", err)
	}
```

- How to register the consumer to subscribe to event messages

see `example/service.order/consumer.go`
see `example/service.notify/consumer.go`

- When cron will run to enable consumer

see `example/cmd/api/main.go`

```
	// start mq consumer
	go func() {
		consume := mq.NeWConsume(pool, logger)
		consume.Consume()
	}()
```

# Increase consume speed

If current consume speed is not satisfying, try alternative approach to rewrite
`consume.go Consume()` to increase performance. Remember using query `for update skip locked`,
so multiple goroutines or multiple pods (from kurbernetes) may consume messages concurrently.

```
func cron() {
	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ticker.C:
			go process()
		}
	}
}

func process() {
    const BatchProcessSize = 5
	capCh := make(chan struct{}, BatchProcessSize)

    // retrive limited number of messages to be consumed from table `queues`
    // query = select id from queues where is_dead = false and check_at < $1 limit 500
    // get queue ids to be consumed
    ids := []int64{}

	for _, id := range ids {
		capCh <- struct{}{}
		go func(queueID int64) {
			defer func() {
				<-capCh
			}()

            // get queue with lock from queue id, but skip locked
			c.consumeSingleMessage(queueID)
		}(id)
	}
}
```
