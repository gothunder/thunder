package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type BareServer struct {
	Server *grpc.Server
}

type NewServerParams struct {
	fx.In

	Logger                *zerolog.Logger
	Interceptors          []grpc.UnaryServerInterceptor `group:"interceptors"`
	MaxReceiveMessageSize *int                          `name:"max_receive_message_size" optional:"true"`
}

func NewServer(params NewServerParams) *BareServer {
	grpcServer := &BareServer{}

	// We want to add the MetadataPropagator interceptor first and
	// logger interceptor last.
	params.Interceptors = append(
		[]grpc.UnaryServerInterceptor{UnaryServerMetadataPropagator},
		append(params.Interceptors, grpcLoggerInterceptor(params.Logger))...,
	)

	// default max message size is 4MB
	maxReceiveMessageSize := 4 * 1024 * 1024
	if params.MaxReceiveMessageSize != nil {
		maxReceiveMessageSize = *params.MaxReceiveMessageSize
	}

	sv := grpc.NewServer(
		grpc.StatsHandler(
			otelgrpc.NewServerHandler(),
		),
		grpc.ChainUnaryInterceptor(
			params.Interceptors...,
		),
		grpc.MaxRecvMsgSize(maxReceiveMessageSize),
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
