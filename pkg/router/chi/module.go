package chi

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	thunderChi "github.com/gothunder/thunder/internal/router/chi"
	"github.com/gothunder/thunder/pkg/router"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		thunderChi.NewRouter,
	),
)

func startListener(lc fx.Lifecycle, s fx.Shutdowner, logger *zerolog.Logger, params router.Params, r *chi.Mux) {
	server, listener, err := thunderChi.CreateServer(params.Handlers, logger, r)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create server")
		err = s.Shutdown()
		if err != nil {
			logger.Error().Err(err).Msg("failed to shutdown")
		}
	}

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				logger.Info().Msg("starting server")

				go func() {
					err := server.Serve(*listener)
					if err != nil && !eris.Is(err, http.ErrServerClosed) {
						logger.Error().Err(err).Msg("error serving requests")
						err = s.Shutdown()
						if err != nil {
							logger.Error().Err(err).Msg("failed to shutdown")
						}
					}
				}()

				return nil
			},
			OnStop: func(ctx context.Context) error {
				logger.Info().Msg("stopping server")

				// This already closes the listener
				err := server.Shutdown(ctx)
				if err != nil {
					logger.Error().Err(err).Msg("error shutting down server")
					return err
				}

				logger.Info().Msg("server stopped")
				return nil
			},
		},
	)
}

var StartServer = fx.Invoke(
	startListener,
)
