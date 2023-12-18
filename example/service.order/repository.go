package order

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Repository is for DB related operations
type Repository interface {
	// create one order, with pending_payment status
	Create(ctx context.Context, tx pgx.Tx, req *createRequest) (orderID int64, err error)

	// cancel one order -- from pending payment status
	Cancel(ctx context.Context, tx pgx.Tx, orderID int64) error

	// get order status with lock
	GetStatusWithLock(ctx context.Context, tx pgx.Tx, orderID int64) (OrderStatus, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool}
}

func (r *repository) Create(ctx context.Context, tx pgx.Tx, req *createRequest) (orderID int64, err error) {
	query := `insert into orders (product, status) values
			($1, $2)
			returning id`
	if err := tx.QueryRow(ctx, query, req.Product, OrderStatusPendingPayment).Scan(&orderID); err != nil {
		return 0, fmt.Errorf("error creating order: %w", err)
	}
	return orderID, nil
}

func (r *repository) Cancel(ctx context.Context, tx pgx.Tx, orderID int64) error {
	query := `update orders set status = $1 where id = $2`
	if _, err := tx.Exec(ctx, query, orderID); err != nil {
		return fmt.Errorf("error canceling order: %w", err)
	}
	return nil
}

func (r *repository) GetStatusWithLock(ctx context.Context, tx pgx.Tx, orderID int64) (status OrderStatus, err error) {
	query := `select status from orders where id = $1 for update`
	if err := tx.QueryRow(ctx, query, orderID).Scan(&status); err != nil {
		return status, fmt.Errorf("error selecting order status: %w", err)
	}
	return "", nil
}
