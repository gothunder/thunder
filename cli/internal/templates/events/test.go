package events

type ExampleTopic string

type ExamplePayload struct {
	Hello string
	World string
}

type ExampleEvent struct {
	Topic   ExampleTopic
	Payload ExamplePayload
}
