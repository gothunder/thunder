package grpc

import (
	"context"
	"net"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func StartTestGrpcService(lc fx.Lifecycle, server GrpcServer, logger *zerolog.Logger) *grpc.ClientConn {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	baseServer := server.GetGrpcServer()
	go func() {
		if err := baseServer.Serve(lis); err != nil {
			logger.Printf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(context.Background(), "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Printf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
		baseServer.Stop()
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			closer()
			return nil
		},
	})

	return conn
}
