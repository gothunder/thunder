package grpc

import (
	"context"
	"net"

	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type GrpcServer interface {
	GetListener() (net.Listener, error)
	GetGrpcServer() *grpc.Server
}

func StartGrpcListener(lc fx.Lifecycle, s fx.Shutdowner, logger *zerolog.Logger, grpcServer GrpcServer) {

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				logger.Info().Msg("starting server")

				go func() {
					lis, err := grpcServer.GetListener()
					if err != nil {
						logger.Error().Err(err).Msg("error getting grpc listener")
						err = s.Shutdown()
						if err != nil {
							logger.Error().Err(err).Msg("failed to shutdown")
						}
					}

					err = grpcServer.GetGrpcServer().Serve(lis)
					if err != nil {
						logger.Error().Err(err).Msg("error serving grpc requests")
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
				grpcServer.GetGrpcServer().Stop()

				logger.Info().Msg("server stopped")
				return nil
			},
		},
	)
}
