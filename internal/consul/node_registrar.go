package consul

import (
	"context"
	"fmt"
	"net"
	"strconv"

	kratosregistry "github.com/go-kratos/kratos/v2/registry"
	"github.com/hashicorp/consul/api"
	"videoCluster/internal/node/conf"
)

// NodeRegistrar registers a Node with the configured Consul agent. Consul
// performs the standard gRPC health check against the advertised address.
type NodeRegistrar struct {
	client *api.Client
	server *conf.Server
}

var _ kratosregistry.Registrar = (*NodeRegistrar)(nil)

func NewNodeRegistrar(client *api.Client, server *conf.Server) (*NodeRegistrar, error) {
	if server == nil || server.AdvertiseHost == "" {
		return nil, fmt.Errorf("server.advertiseHost is required for Consul registration")
	}
	if _, err := portFromAddress(server.Grpc.GetAddr()); err != nil {
		return nil, fmt.Errorf("invalid server.grpc.addr: %w", err)
	}
	if _, err := portFromAddress(server.Http.GetAddr()); err != nil {
		return nil, fmt.Errorf("invalid server.http.addr: %w", err)
	}
	return &NodeRegistrar{client: client, server: server}, nil
}

func (r *NodeRegistrar) Register(ctx context.Context, service *kratosregistry.ServiceInstance) error {
	grpcPort, err := portFromAddress(r.server.Grpc.GetAddr())
	if err != nil {
		return err
	}
	httpPort, err := portFromAddress(r.server.Http.GetAddr())
	if err != nil {
		return err
	}

	registration := &api.AgentServiceRegistration{
		ID:      service.ID,
		Name:    service.Name,
		Address: r.server.AdvertiseHost,
		Port:    grpcPort,
		Meta: map[string]string{
			"grpc_port": strconv.Itoa(grpcPort),
			"http_port": strconv.Itoa(httpPort),
		},
		TaggedAddresses: map[string]api.ServiceAddress{
			"grpc": {Address: r.server.AdvertiseHost, Port: grpcPort},
			"http": {Address: r.server.AdvertiseHost, Port: httpPort},
		},
		Check: &api.AgentServiceCheck{
			Name:                           "gRPC health",
			GRPC:                           net.JoinHostPort(r.server.AdvertiseHost, strconv.Itoa(grpcPort)),
			Interval:                       "10s",
			Timeout:                        "2s",
			DeregisterCriticalServiceAfter: "1m",
		},
	}
	return r.client.Agent().ServiceRegisterOpts(registration, api.ServiceRegisterOpts{
		ReplaceExistingChecks: true,
	}.WithContext(ctx))
}

func (r *NodeRegistrar) Deregister(ctx context.Context, service *kratosregistry.ServiceInstance) error {
	return r.client.Agent().ServiceDeregisterOpts(service.ID, (&api.QueryOptions{}).WithContext(ctx))
}

func portFromAddress(address string) (int, error) {
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return 0, err
	}
	value, err := strconv.Atoi(port)
	if err != nil || value < 1 || value > 65535 {
		return 0, fmt.Errorf("invalid port %q", port)
	}
	return value, nil
}
