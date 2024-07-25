package events

import "go.uber.org/fx"

// NamedHandler is a handler that has a queue name associated with it.
type NamedHandler interface {
	QueuePosfix() string
	Handler
}

type namedHandler struct {
	Handler
	queuePosfix string
}

func (n *namedHandler) QueuePosfix() string {
	return n.queuePosfix
}

// NewNamedHandlerFromHandler creates a new NamedHandler from a Handler.
func NewNamedHandlerFromHandler(handler Handler, queuePosfix string) NamedHandler {
	return &namedHandler{
		Handler:     handler,
		queuePosfix: queuePosfix,
	}
}

// FxAnnotateNamedHandler provides a named handler with group tags.
// This is used to provide named handlers with the group tag `group:"named_handlers"`.
// It's not a type safe function, so it's up to the caller to provide the correct type.
func FxAnnotateNamedHandler(namedHandlerFunc interface{}) fx.Option {
	return fx.Provide(
		fx.Annotate(namedHandlerFunc, fx.As(new(NamedHandler)), fx.ResultTags(`group:"named_handlers"`)),
	)
}
