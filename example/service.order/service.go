package order

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/smiletrl/mq"
)

type Service interface {
	// create one order, with pending_payment status
	Create(ctx context.Context, req *createRequest) error
}

type service struct {
	pool *pgxpool.Pool
	repo Repository
	mq   mq.Provider
}

func NewService(repository Repository) Service {
	return &service{
		repo: repository,
	}
}

func (s *service) Create(ctx context.Context, req *createRequest) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	orderID, err := s.repo.Create(ctx, tx, req)
	if err != nil {
		return fmt.Errorf("error creating order at service: %w", err)
	}

	// send order created mq message
	msg := mq.NewOrderMessage(mq.EventOrderCreated).
		WithOrderID(orderID).
		Encode(ctx)
	if err := s.mq.SendMessage(ctx, tx, msg); err != nil {
		return fmt.Errorf("error sending mq message: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing: %w", err)
	}

	return nil
}
