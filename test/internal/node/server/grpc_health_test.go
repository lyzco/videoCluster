package server_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	stdgrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	v1 "videoCluster/api/node/v1"
	"videoCluster/internal/node/conf"
	"videoCluster/internal/node/server"
	"videoCluster/internal/node/service"
)

func TestNodeGRPCServerServesStandardHealthProtocol(t *testing.T) {
	srv := server.NewGRPCServer(&conf.Server{
		Grpc: &conf.Server_GRPC{Addr: "127.0.0.1:0"},
	}, &service.VideoService{UnimplementedNodeServiceServer: v1.UnimplementedNodeServiceServer{}}, log.NewStdLogger(nil))

	endpoint, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = srv.Start(ctx) }()
	t.Cleanup(func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), time.Second)
		defer stopCancel()
		_ = srv.Stop(stopCtx)
	})

	connection, err := stdgrpc.DialContext(context.Background(), grpcAddress(t, endpoint), stdgrpc.WithTransportCredentials(insecure.NewCredentials()), stdgrpc.WithBlock())
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()

	response, err := grpc_health_v1.NewHealthClient(connection).Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("health status = %s, want SERVING", response.Status)
	}
}

func grpcAddress(t *testing.T, endpoint *url.URL) string {
	t.Helper()
	if endpoint.Host == "" {
		t.Fatalf("endpoint has no host: %s", endpoint)
	}
	return endpoint.Host
}
