package outbox

import (
	"sort"

	"github.com/TheRafaBonin/roxy"
	internaloutbox "github.com/gothunder/thunder/internal/events/outbox"
)

type Storer interface {
	internaloutbox.Storer
}

type TransactionalStorer interface {
	internaloutbox.TransactionalStorer
}

func NewOutboxStorer(options ...StorerOptions) (storer Storer, err error) {
	cfg := &config{
		wrappers: make([]storerWrapper, 0),
	}
	for _, opt := range options {
		opt(cfg)
	}
	sort.Slice(cfg.wrappers, func(i, j int) bool {
		return cfg.wrappers[i].order < cfg.wrappers[j].order
	})

	storer = internaloutbox.NewStorer()
	for _, wrapper := range cfg.wrappers {
		storer, err = wrapper.wrap(storer)
		if err != nil {
			return nil, roxy.Wrap(err, "wrapping storer")
		}
	}
	return storer, nil
}

type config struct {
	wrappers []storerWrapper
}

type storerWrapper struct {
	wrap  func(Storer) (Storer, error)
	order int // the higher the number, the more external the wrapper
}

type StorerOptions func(*config)

func WithTracing() StorerOptions {
	return func(cfg *config) {
		wrapper := storerWrapper{
			wrap: func(storer Storer) (Storer, error) {
				return internaloutbox.WrapStorerWithTracing(storer), nil
			},
			// tracing should be the most external wrapper
			// so it initializes the tracing span
			// and makes it acessible to the other wrappers
			order: 1,
		}
		cfg.wrappers = append(cfg.wrappers, wrapper)
	}
}

func WithLogging() StorerOptions {
	return func(cfg *config) {
		wrapper := storerWrapper{
			wrap: func(storer Storer) (Storer, error) {
				return internaloutbox.WrapStorerWithLogging(storer), nil
			},
			order: 0,
		}
		cfg.wrappers = append(cfg.wrappers, wrapper)
	}
}

func WithMetrics() StorerOptions {
	return func(cfg *config) {
		wrapper := storerWrapper{
			wrap: func(storer Storer) (Storer, error) {
				return internaloutbox.WrapStorerWithMetrics(storer)
			},
			order: 0,
		}
		cfg.wrappers = append(cfg.wrappers, wrapper)
	}
}
