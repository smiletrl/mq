package mq

// create one new order message
func NewOrderMessage(e Event) OrderMessage {
	return &MQMessage{
		MQEvent: e,
	}
}

// Order Message
type OrderMessage interface {
	Message

	// set order id
	WithOrderID(orderID int64) OrderMessage

	// get order id
	OrderID() int64
}

// convert struct to message interface to be used by consumer
func (m *MQMessage) OrderMessage() OrderMessage {
	return m
}

func (m *MQMessage) WithOrderID(orderID int64) OrderMessage {
	m.MQOrderID = orderID
	return m
}

func (m *MQMessage) OrderID() int64 {
	return m.MQOrderID
}
