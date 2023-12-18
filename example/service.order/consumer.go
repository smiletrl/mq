package order

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/smiletrl/mq"
)

func RegisterConsumer(repo Repository) {
	mq.RegisterConsumer(&OrderCreatedConsumer{
		repo: repo,
	})
}

// Consume when order is created. This is primarily to cancel the order if order is created but not paid.
type OrderCreatedConsumer struct {
	repo Repository
}

func (c *OrderCreatedConsumer) Name() string {
	return "order:order:order_created"
}

func (c *OrderCreatedConsumer) Event() mq.Event {
	return mq.EventOrderCreated
}

func (c *OrderCreatedConsumer) Delay() time.Duration {
	// wait 5 minutes to consume this message
	return 5 * time.Minute
}

func (c *OrderCreatedConsumer) Consume(ctx context.Context, tx pgx.Tx, msgRaw *mq.MQMessage) error {
	msg := msgRaw.OrderMessage()

	ctx = mq.NewContext(ctx, msg)
	orderID := msg.OrderID()

	// get order status
	status, err := c.repo.GetStatusWithLock(ctx, tx, orderID)
	if err != nil {
		return fmt.Errorf("error getting status with lock at consume: %w", err)
	}

	// if this order status is not pending payment, then no need to process this message
	if status != OrderStatusPendingPayment {
		return nil
	}

	// if this order is still pending payment, then cancel the order
	if err := c.repo.Cancel(ctx, tx, orderID); err != nil {
		return fmt.Errorf("error canceling order at consumer: %w", err)
	}

	return nil
}
