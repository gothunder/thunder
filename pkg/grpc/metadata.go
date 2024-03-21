package grpc

import (
	"context"
	"strings"

	thunderContext "github.com/gothunder/thunder/pkg/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryServerMetadataPropagator(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	grpcMd, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req)
	}

	md := make(thunderContext.Metadata, len(grpcMd))
	for k, v := range grpcMd {
		if len(v) == 1 {
			md.Set(k, v[0])
		} else if len(v) > 1 {
			md.Set(k, strings.Join(v, ","))
		}
	}
	ctx = thunderContext.ContextWithMetadata(ctx, md)
	return handler(ctx, req)
}

func UnaryClientMetadataPropagator(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md := thunderContext.MetadataFromContext(ctx)
	if md == nil {
		invoker(ctx, method, req, reply, cc, opts...)
	}

	grpcMd := metadata.New(md)
	for k, v := range md {
		grpcMd.Set(k, v)
	}
	ctx = metadata.NewOutgoingContext(ctx, grpcMd)
	return invoker(ctx, method, req, reply, cc, opts...)
}
