package mq

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/gommon/log"
)

const (
	ConsumerNotFound = "consumer not found"
	MaxRetry         = 5
)

var (
	// retry message consume
	RetryDelay = map[int]time.Duration{
		0: 2 * time.Second,
		1: 10 * time.Second,
		2: time.Minute,
		3: 30 * time.Minute,
		4: time.Hour,
	}
)

type Consume interface {
	Consume()
}

type consume struct {
	pool *pgxpool.Pool
}

func NeWConsume(tenantPool *pgxpool.Pool) Consume {
	return consume{tenantPool}
}

type Queue struct {
	ID int64
	// consumer name
	ConsumerName string
	Message      MQMessage
	Retry        int
	IsDead       bool
	FailedReason *string
	CheckAt      time.Time
	CreatedAT    time.Time
}

func (c consume) Consume() {
	fmt.Println("consume started")

	rand.Seed(time.Now().UnixNano())

	for {
		if sleep := c.consumeSingleMessage(); sleep {
			// sleep between 0-3 seconds
			r := rand.Intn(3)
			time.Sleep(time.Duration(r) * time.Second)
		}
	}
}

func (c consume) consumeSingleMessage() (sleep bool) {
	// catch possible panic
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("MQ: panic: %+v", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	tx, err := c.pool.Begin(ctx)
	if err != nil {
		log.Errorf("MQ: error starting tx at consume: (%v)", err)
		sleep = true
		return
	}
	defer tx.Rollback(ctx)

	// get the single queue
	queues := []Queue{}
	query := `select * from queues where is_dead = false and check_at < $1 limit 1 for update skip locked`

	if err := pgxscan.Select(ctx, tx, &queues, query, time.Now()); err != nil {
		log.Errorf("MQ: error selecting message at consume: (%v)", err)
		return
	}
	if len(queues) == 0 {
		sleep = true
		return
	}

	// get this message's consumer
	queue := queues[0]
	consumer, ok := consumers[queue.ConsumerName]
	if !ok {
		log.Errorf("MQ: consumer is not found: %s", queue.ConsumerName)

		sql := "update queues set is_dead = true, failed_reason = $1 where id = $2"
		if _, err := tx.Exec(ctx, sql, ConsumerNotFound, queue.ID); err != nil {
			log.Errorf("MQ: error setting message to be dead: %w", err)
		}
		// commit tx
		if err := tx.Commit(ctx); err != nil {
			log.Errorf("MQ: error committing tx: %w", err)
		}
		return
	}

	// consumer will use a nested transaction, so if the nested transaction within the consumer has
	// failed, we can still log the retry or failed reason in the outer transaction.
	nestTx, err := tx.Begin(ctx)
	if err != nil {
		log.Errorf("MQ: error begining nested tx: %w", err)
		return
	}

	// consume this message
	if consumeErr := consumer.Consume(ctx, nestTx, &queue.Message); consumeErr != nil {
		log.Errorf("MQ: consume message failed with error: (%v)", consumeErr)

		// rollback the nested transaction
		if err := nestTx.Rollback(ctx); err != nil {
			// if this error happens, something fatal happens, this message will be processed infinitely.
			log.Errorf("MQ: error rolling back nested tx")
			return
		}
		if queue.Retry+1 >= MaxRetry {
			// max retry reached, set this message to be dead
			sql := `update queues set retry = retry + 1, is_dead = true, failed_reason = $1 where id = $2`
			if _, err := tx.Exec(ctx, sql, consumeErr.Error(), queue.ID); err != nil {
				log.Errorf("MQ: error updating queues with retry and is_dead: (%v)", err)
			}
		} else {
			// increase this retry, and check it later
			sql := `update queues set retry = retry + 1, check_at = $1 where id = $2`
			delay, ok := RetryDelay[queue.Retry]
			if !ok {
				delay = 10 * time.Second
			}
			if _, err := tx.Exec(ctx, sql, time.Now().Add(delay), queue.ID); err != nil {
				log.Errorf("MQ: error updating queues with retry: (%v)", err)
			}
		}
		// commit tx
		if err := tx.Commit(ctx); err != nil {
			log.Errorf("MQ: error committing tx")
		}
		return
	}

	// delete this message if all goes well
	sql := `delete from queues where id = $1`
	if _, err := tx.Exec(ctx, sql, queue.ID); err != nil {
		log.Errorf("MQ: delete queue with error: (%v)", err)
		return
	}

	// commit tx
	if err := tx.Commit(ctx); err != nil {
		log.Errorf("MQ: error committing tx")
	}
	return sleep
}
