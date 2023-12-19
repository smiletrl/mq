package mq

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// message keys
const (
	RequestID string = "request_id"
)

// RocketMessage reprents one message used to encode and decode custom information.
type Message interface {
	// encode the message to be saved into postgres table
	Encode(ctx context.Context) Message

	// get event
	Event() Event

	// get request id
	RequestID() string
}

// MQMessage holds message data, which will be saved at db table
type MQMessage struct {
	// event
	MQEvent Event `json:"event"`

	// request id
	MQRequestID string `json:"request_id,omitempty"`

	// order id, optional
	MQOrderID int64 `json:"order_id,omitempty"`
}

func (m MQMessage) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *MQMessage) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, m)
}

func (m *MQMessage) Encode(ctx context.Context) Message {
	if ctx.Value(RequestID) != nil {
		reqID := ctx.Value(RequestID).(string)
		m.MQRequestID = reqID
	}

	return m
}

func (m *MQMessage) Event() Event {
	return m.MQEvent
}

func (m *MQMessage) WithRequestID(requestID string) OrderMessage {
	m.MQRequestID = requestID
	return m
}

func (m *MQMessage) RequestID() string {
	return m.MQRequestID
}
