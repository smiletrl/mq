package notify

import (
	"context"
)

type Service interface {
	SendEmail(ctx context.Context, orderID int64) error
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) SendEmail(ctx context.Context, orderID int64) error {
	// invoke email service and send email notification for this order
	return nil
}
