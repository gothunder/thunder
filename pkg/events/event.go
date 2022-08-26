package events

type Event struct {
	// The name of the event.
	Topic string

	// The payload of the event.
	Payload interface{}
}
