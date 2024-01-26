package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/rs/zerolog"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type BareServer struct {
	Server *grpc.Server
}

type NewServerParams struct {
	fx.In

	Logger       *zerolog.Logger
	Interceptors []grpc.UnaryServerInterceptor `group:"interceptors"`
}

func NewServer(params NewServerParams) *BareServer {
	grpcServer := &BareServer{}

	params.Interceptors = append(
		params.Interceptors,
		grpcLoggerInterceptor(params.Logger),
	)

	sv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			params.Interceptors...,
		),
	)
	grpcServer.Server = sv

	return grpcServer
}

func (s *BareServer) GetListener() (net.Listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50051))
	if err != nil {
		return nil, err
	}

	return lis, nil
}

func (s *BareServer) GetGrpcServer() *grpc.Server {
	return s.Server
}

func grpcLoggerInterceptor(logger *zerolog.Logger) func(context.Context, interface{}, *grpc.UnaryServerInfo, grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Add logger to context
		lctx := logger.WithContext(ctx)

		// Calls the handler
		h, err := handler(lctx, req)

		return h, err
	}
}
