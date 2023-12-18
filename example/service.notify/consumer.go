package notify

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/smiletrl/mq"
)

func RegisterConsumer(service Service) {
	mq.RegisterConsumer(&OrderCreatedConsumer{
		svc: service,
	})
}

// Consume when order is created. This is primarily to send one order email message out.
type OrderCreatedConsumer struct {
	svc Service
}

func (c *OrderCreatedConsumer) Name() string {
	return "notify:order:order_created"
}

func (c *OrderCreatedConsumer) Event() mq.Event {
	return mq.EventOrderCreated
}

func (c *OrderCreatedConsumer) Delay() time.Duration {
	// consume this message immediately
	return 0
}

func (c *OrderCreatedConsumer) Consume(ctx context.Context, tx pgx.Tx, msgRaw *mq.MQMessage) error {
	msg := msgRaw.OrderMessage()

	ctx = mq.NewContext(ctx, msg)
	orderID := msg.OrderID()

	// send email notification for this order.
	if err := c.svc.SendEmail(ctx, orderID); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	return nil
}
