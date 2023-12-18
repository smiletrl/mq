package mq

import (
	"context"
)

func NewContext(ctx context.Context, msg Message) context.Context {
	// http request middleware may add request id to track mq messages
	// In case something wrong, request id is helpful to find the original http request
	// and debug potential error.
	ctx = context.WithValue(ctx, RequestID, msg.RequestID())
	return ctx
}
