package order

type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "pending_payment"
	OrderStatusCanceled       OrderStatus = "canceled"
)
