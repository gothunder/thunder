package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// TestSuppliedInterceptorEnrichmentReachesHandler is the regression for the
// interceptor-ordering bug: the base logger interceptor used to run AFTER
// supplied interceptors, replacing the context logger and silently dropping
// any enrichment (trace_id, audit fields) before the handler executed.
//
// It simulates a supplied interceptor that enriches the context logger (the
// same pattern backend-commons logs.UnaryServerInterceptor uses) and asserts
// the handler's log line actually carries the enriched fields.
func TestSuppliedInterceptorEnrichmentReachesHandler(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	// Supplied interceptor: enriches the context logger, exactly like
	// backend-commons logs.UnaryServerInterceptor does for trace_id.
	enriching := func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		l := zerolog.Ctx(ctx).With().
			Str("trace_id", "trace-abc-123").
			Str("grpc_method", info.FullMethod).
			Logger()
		return handler(l.WithContext(ctx), req)
	}

	// Compose the chain the same way NewServer does.
	interceptors := append(
		[]grpc.UnaryServerInterceptor{
			UnaryServerMetadataPropagator,
			grpcLoggerInterceptor(&logger),
		},
		enriching,
	)

	// Handler logs from its context, like real service handlers do.
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		zerolog.Ctx(ctx).Info().Msg("handled")
		return "ok", nil
	}

	// Manually chain (mirrors grpc.ChainUnaryInterceptor semantics).
	chained := handler
	for i := len(interceptors) - 1; i >= 0; i-- {
		ic := interceptors[i]
		next := chained
		info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Do"}
		chained = func(ctx context.Context, req interface{}) (interface{}, error) {
			return ic(ctx, req, info, next)
		}
	}

	if _, err := chained(context.Background(), "req"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse handler log: %v\noutput: %s", err, buf.String())
	}

	if entry["trace_id"] != "trace-abc-123" {
		t.Errorf("handler log trace_id = %v, want trace-abc-123 (enrichment was dropped)", entry["trace_id"])
	}
	if entry["grpc_method"] != "/test.Service/Do" {
		t.Errorf("handler log grpc_method = %v, want /test.Service/Do", entry["grpc_method"])
	}
}
