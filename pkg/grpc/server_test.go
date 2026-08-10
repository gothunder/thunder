package grpc

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

// chain replicates grpc.ChainUnaryInterceptor semantics: interceptors run in
// slice order, each wrapping the next, with the handler innermost.
func chain(interceptors []grpc.UnaryServerInterceptor, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) grpc.UnaryHandler {
	chained := handler
	for i := len(interceptors) - 1; i >= 0; i-- {
		ic := interceptors[i]
		next := chained
		chained = func(ctx context.Context, req interface{}) (interface{}, error) {
			return ic(ctx, req, info, next)
		}
	}
	return chained
}

// TestComposedChainPreservesSuppliedEnrichment is the regression for the
// interceptor-ordering bug: the base logger interceptor used to be appended
// AFTER supplied interceptors, replacing the context logger and silently
// dropping their enrichment (trace_id, audit fields) before the handler ran.
//
// Unlike a hand-assembled chain, this test takes the interceptor slice from
// composeInterceptors — the same production function NewServer uses — so
// reverting the ordering there makes this test fail.
func TestComposedChainPreservesSuppliedEnrichment(t *testing.T) {
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

	// Compose through the PRODUCTION function used by NewServer.
	interceptors := composeInterceptors(&logger, []grpc.UnaryServerInterceptor{enriching})

	// Handler logs from its context, like real service handlers do.
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		zerolog.Ctx(ctx).Info().Msg("handled")
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Do"}
	if _, err := chain(interceptors, info, handler)(context.Background(), "req"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse handler log: %v\noutput: %s", err, buf.String())
	}

	if entry["trace_id"] != "trace-abc-123" {
		t.Errorf("handler log trace_id = %v, want trace-abc-123 (supplied enrichment was dropped)", entry["trace_id"])
	}
	if entry["grpc_method"] != "/test.Service/Do" {
		t.Errorf("handler log grpc_method = %v, want /test.Service/Do", entry["grpc_method"])
	}
}

// TestComposedChainBaseLoggerReachesHandlerWithoutSupplied guards the base
// behavior: with no supplied interceptors, the handler still gets the base
// logger in its context.
func TestComposedChainBaseLoggerReachesHandlerWithoutSupplied(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	interceptors := composeInterceptors(&logger, nil)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		zerolog.Ctx(ctx).Info().Str("marker", "base").Msg("handled")
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Do"}
	if _, err := chain(interceptors, info, handler)(context.Background(), "req"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Contains(buf.Bytes(), []byte(`"marker":"base"`)) {
		t.Errorf("handler did not log through the base logger: %s", buf.String())
	}
}
