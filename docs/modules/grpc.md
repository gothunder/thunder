# GRPC

## Prerequisites

Install the prerequisites listed in [gRPC Go - Quick
Start](https://grpc.io/docs/languages/go/quickstart/#prerequisites).

## Getting Started

In main.go:
```go
    ...
	fx.New(
		fx.Populate(&w),

		// Adapters
		thunderRabbitMQ.PublisherModule,
		commonsConnection.Native,
		thunderRouter.Module,
		thunderLog.Module,
		thunderGraphql.Module,

		// Internal Modules
		transportoutbound.Module,
		transportinbound.Module,
		generated.Module,
		features.Module,

		// Listeners
		fx.Invoke(
			commonsMigrate.MigrateUp,
			auth.RegisterMiddleware,
			thunderRouter.StartListener,
			thunderGrpc.StartGrpcListener,   <----
		),
	).Run()
    ...
```

Inside transport-inbound module provide the gRPC server:
```go
type server struct {
	pb.UnimplementedSuperadminServer
	grpcServer *grpc.Server
	Queries    *queries.QueryGroup
}

func NewGrpcServer(queries *queries.QueryGroup) thunderGrpc.GrpcServer {
	serv := server{
		Queries: queries,
	}
	sv := grpc.NewServer()
	serv.grpcServer = sv
	pb.RegisterSuperadminServer(sv, serv)
	return serv
}

func (s server) GetListener() (net.Listener, error) {   <-------
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50051))
	if err != nil {
		return nil, err
	}

	return lis, nil
}

func (s server) GetGrpcServer() *grpc.Server {   <-------
	return s.grpcServer
}
```
