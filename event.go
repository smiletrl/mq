package mq

// event
type Event string

func (e Event) String() string {
	return string(e)
}

const (
	// event when order is created
	EventOrderCreated Event = "order_created"
)
