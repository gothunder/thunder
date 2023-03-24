# Introduction to Thunder

Thunder is a collection of libraries and opinionated patterns to build
cloud-native services. The project provides different modules, which can be
used individually and replaced at any time.

Most of the modules use consolidated projects under the hood, the idea of this
project is to provide wrappers and connect all of these pieces.

## How it works or Features

This project provides modules and constructors for all of its components, so
one may use a dependency injection framework such as [Uber's
fx](https://uber-go.github.io/fx/), which is used in the docs, or manually
instantiate the components.

List of modules:

- log: Provides a logger through [zerolog](https://github.com/rs/zerolog).
- graphql: Provides a GraphQL handler using [gqlgen](https://gqlgen.com/).
- chi: Provides a [chi](https://go-chi.io/#/) multiplexer with the default
       middlewares. Also exposes a method to start a server with graceful
       shutdown.
- mocks: Provides mocks for consumer and publisher, along with a channel to
         receive intercepted messages.
- rabbitmq: Provides a RabbitMQ publisher and consumer with
            [amqp]("https://github.com/rabbitmq/amqp091-go").

## Getting started

The modules are imported to the root of your app, you can then provide them to
the rest of your project, below we're using [Uber's
fx](https://uber-go.github.io/fx/) to perform the dependency injection

```go
// main.go
package main

import (
    transportinbound "github.com/gothunder/thunder/example/internal/transport-inbound"
    transportoutbound "github.com/gothunder/thunder/example/internal/transport-outbound"
    thunderEventRabbitmq "github.com/gothunder/thunder/pkg/events/rabbitmq"
    thunderLogs "github.com/gothunder/thunder/pkg/log"

    "github.com/rs/zerolog/diode"
    "go.uber.org/fx"
)

func main() {
    var w diode.Writer

    app := fx.New(
        // The order of these options isn't important.
        thunderLogs.Module,
        fx.Populate(&w),

        transportinbound.Module,
        transportoutbound.Module,

        thunderEventRabbitmq.PublisherModule,
        thunderEventRabbitmq.InvokeConsumer,
    )
    app.Run()

    // This is required to flush the logs to stdout.
    // We only want to do this after the app has exited.
    thunderLogs.DiodeShutdown(w)
}
```
