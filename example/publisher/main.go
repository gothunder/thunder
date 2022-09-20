package main

import (
	"context"
	"fmt"

	"github.com/gothunder/thunder/pkg/events"
	"github.com/gothunder/thunder/pkg/events/rabbitmq"
	"github.com/gothunder/thunder/pkg/log"
	"go.uber.org/fx"
)

type testEvent struct {
	Hello string `json:"hello"`
}

func main() {
	var publisher events.EventPublisher
	app := fx.New(
		fx.Populate(&publisher),
		log.Module,
		rabbitmq.PublisherModule,
	)
	go app.Run()

	for i := 0; i < 1000; i++ {
		publisher.Publish(context.Background(), events.Event{
			Topic:   "topic.test",
			Payload: testEvent{Hello: fmt.Sprintf("world, %d", i)},
		})
	}

	err := app.Stop(context.Background())
	if err != nil {
		panic(err)
	}
}
