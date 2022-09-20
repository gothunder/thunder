package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
		fmt.Printf("Received event: %s \n", topic)

		event := testEvent{}
		err := json.Unmarshal(payload, &event)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Payload: %v \n", event)

		time.Sleep(200 * time.Millisecond)
		fmt.Println("Done")

		return events.Success
	}

	app := fx.New(
		log.Module,
		rabbitmq.InvokeConsumer([]string{"topic.test"}, handler),
	)
	app.Run()
}
