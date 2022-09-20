package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gothunder/thunder/pkg/events"
	"github.com/gothunder/thunder/pkg/events/rabbitmq"
	"github.com/gothunder/thunder/pkg/log"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

type testEvent struct {
	Hello string `json:"hello"`
}

func main() {
	var publisher events.EventPublisher
	var logger *zerolog.Logger
	app := fx.New(
		fx.Populate(&publisher, &logger),
		log.Module,
		rabbitmq.PublisherModule,
	)
	go app.Run()

	ctx := logger.WithContext(context.Background())
	for i := 0; i < 20; i++ {
		err := publisher.Publish(ctx, events.Event{
			Topic:   "topic.test",
			Payload: testEvent{Hello: fmt.Sprintf("world, %d", i)},
		})
		if err != nil {
			panic(err)
		}

		time.Sleep(1 * time.Second)
	}

	err := app.Stop(ctx)
	if err != nil {
		panic(err)
	}
}
