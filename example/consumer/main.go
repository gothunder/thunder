package main

import (
	"context"
	"encoding/json"

	"github.com/gothunder/thunder/pkg/events"
	"github.com/gothunder/thunder/pkg/events/rabbitmq"
	"github.com/gothunder/thunder/pkg/log"
	"go.uber.org/fx"
)

type testEvent struct {
	Hello string `json:"hello"`
}

func main() {
	handler := func(ctx context.Context, topic string, payload []byte) events.HandlerResponse {
		event := testEvent{}
		err := json.Unmarshal(payload, &event)
		if err != nil {
			panic(err)
		}

		return events.Success
	}

	app := fx.New(
		log.Module,
		rabbitmq.InvokeConsumer([]string{"topic.test"}, handler),
	)
	app.Run()
}
