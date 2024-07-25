package events

import (
	"context"
	"regexp"

	"github.com/TheRafaBonin/roxy"
)

type HandlerResponse int

const (
	// Default, we remove the message from the queue.
	Success HandlerResponse = iota

	// The message will be delivered to a server configured dead-letter queue.
	DeadLetter

	// Deliver this message to a different worker.
	Retry

	RetryBackoff
)

type EventDecoder interface {
	// Decode decodes the payload into the given interface.
	// Returns an error if the payload cannot be decoded.
	Decode(v interface{}) error
}

type Handler interface {
	// The topics that the consumer will be subscribed to.
	Topics() []string

	// The function that will be called when a subscribed message is received.
	Handle(ctx context.Context, topic string, decoder EventDecoder) HandlerResponse
}

/*
MatchTopicAndFormatsMessage matches the topic and formats the message.

**This is deprecated and should not be used.**

It takes the following parameters:
  - ctx: the context.Context object for the function.
  - decoder: the thunderEvents.EventDecoder object for decoding the message.
  - referenceTopic: the reference topic string for matching.
  - topic: the topic string to match against the reference topic.
  - message: the message object to be formatted.

It returns a pointer to the formatted message object and any error encountered.
if the match fails, nil is returned.
*/
func MatchTopicAndFormatsMessage[T any](
	_ context.Context,
	decoder EventDecoder,
	referenceTopic string,
	topic string,
	message T,
) (*T, error) {
	// Declare some variables
	var match bool
	var err error

	match, err = regexp.MatchString(referenceTopic, topic)
	err = roxy.Wrap(err, "Failed to match topic")
	if err != nil {
		return nil, err
	}
	if match {
		err := decoder.Decode(&message)
		roxy.Wrap(err, "unmarshalling message")
		if err != nil {
			return nil, err
		}

		return &message, nil
	}

	return nil, nil
}
