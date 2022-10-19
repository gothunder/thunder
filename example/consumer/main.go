package main

import (
	"context"

	"github.com/gothunder/thunder/pkg/events"
	"github.com/gothunder/thunder/pkg/events/rabbitmq"
	"github.com/gothunder/thunder/pkg/log"
	"github.com/rs/zerolog/diode"
	"go.uber.org/fx"
)

type testEvent struct {
	Hello string `json:"hello"`
}

type testHandler struct{}

func newHandler() events.Handler {
	return testHandler{}
}

func (t testHandler) Topics() []string {
	return []string{
		"topic.test",
	}
}

func (t testHandler) Handle(ctx context.Context, topic string, decoder events.EventDecoder) events.HandlerResponse {
	event := testEvent{}
	err := decoder.Decode(&event)
	if err != nil {
		panic(err)
	}
	if event.Hello == "world, 3" {
		return events.DeadLetter
	}

	return events.Success
}

func main() {
	var w diode.Writer
	app := fx.New(
		fx.Populate(&w),
		log.Module,
		fx.Provide(
			func() events.Handler {
				return newHandler()
			},
		),
		rabbitmq.InvokeConsumer,
	)
	app.Run()

	log.DiodeShutdown(w)
}
