package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gothunder/thunder/pkg/events"
	"github.com/gothunder/thunder/pkg/events/rabbitmq"
	"github.com/gothunder/thunder/pkg/log"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"go.uber.org/fx"
)

type testEvent struct {
	Hello string `json:"hello"`
}

func main() {
	var publisher events.EventPublisher
	var logger *zerolog.Logger
	var w diode.Writer

	app := fx.New(
		fx.Populate(&publisher, &logger, &w),
		log.Module,
		rabbitmq.PublisherModule,
	)
	go func() {
		time.Sleep(5 * time.Second)
		ctx := logger.WithContext(context.Background())
		for i := 0; i < 10; i++ {
			err := publisher.Publish(ctx, "topic.test", testEvent{
				Hello: fmt.Sprintf("world, %d", i),
			})
			if err != nil {
				panic(err)
			}
		}
	}()
	app.Run()

	log.DiodeShutdown(w)
}
