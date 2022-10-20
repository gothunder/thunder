## Thunder

Thunder is a collection of libraries and opinionated patterns to build cloud-native services. The project provides different modules, which can be used individually and replaced at any time.

Most of the modules use consolidated projects under the hood, the idea of this project is to provide wrappers and connect all of these pieces.

The modules are imported to the root of your app, you can then provide them to the rest of your project, below we're using Uber fx to perform the dependency injection.

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

You can check your [docs](https://go-thunder.gitbook.io/) for instructions on how to use it, also there's a basic [example](https://github.com/gothunder/thunder/tree/main/example).