package router

import (
	"net/http"

	"go.uber.org/fx"
)

type HTTPHandler interface {
	Method() string
	Pattern() string
	http.Handler
}

type HandlerOutput struct {
	fx.Out

	Handler HTTPHandler `group:"handlers"`
}

type Params struct {
	fx.In

	Handlers []HTTPHandler `group:"handlers"`
}
