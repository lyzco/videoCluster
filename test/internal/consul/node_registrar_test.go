package consul_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kratosregistry "github.com/go-kratos/kratos/v2/registry"
	"github.com/hashicorp/consul/api"
	"videoCluster/internal/consul"
	"videoCluster/internal/node/conf"
)

func TestNodeRegistrarRegistersAdvertisedGRPCHealthCheck(t *testing.T) {
	var registration api.AgentServiceRegistration
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/v1/agent/service/register" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("replace-existing-checks") != "true" {
			t.Fatal("replace-existing-checks was not set")
		}
		if err := json.NewDecoder(r.Body).Decode(&registration); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := api.DefaultConfig()
	config.Address = strings.TrimPrefix(server.URL, "http://")
	config.HttpClient = server.Client()
	client, err := api.NewClient(config)
	if err != nil {
		t.Fatal(err)
	}

	registrar, err := consul.NewNodeRegistrar(client, &conf.Server{
		AdvertiseHost: "192.168.1.25",
		Grpc:          &conf.Server_GRPC{Addr: "0.0.0.0:9000"},
		Http:          &conf.Server_HTTP{Addr: "0.0.0.0:8000"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := registrar.Register(context.Background(), &kratosregistry.ServiceInstance{
		ID:   "media-node-1",
		Name: "videoCluster-media-node",
	}); err != nil {
		t.Fatal(err)
	}

	if registration.ID != "media-node-1" || registration.Name != "videoCluster-media-node" {
		t.Fatalf("service = %#v", registration)
	}
	if registration.Address != "192.168.1.25" || registration.Port != 9000 {
		t.Fatalf("address/port = %s/%d", registration.Address, registration.Port)
	}
	if registration.Check == nil || registration.Check.GRPC != "192.168.1.25:9000" {
		t.Fatalf("gRPC check = %#v", registration.Check)
	}
	if registration.Check.Interval != "10s" || registration.Check.Timeout != "2s" || registration.Check.DeregisterCriticalServiceAfter != "1m" {
		t.Fatalf("unexpected check timing: %#v", registration.Check)
	}
	if registration.TaggedAddresses["http"].Port != 8000 || registration.TaggedAddresses["grpc"].Port != 9000 {
		t.Fatalf("tagged addresses = %#v", registration.TaggedAddresses)
	}
}

func TestNewNodeRegistrarRequiresAdvertisedAddress(t *testing.T) {
	_, err := consul.NewNodeRegistrar(nil, &conf.Server{
		Grpc: &conf.Server_GRPC{Addr: "0.0.0.0:9000"},
		Http: &conf.Server_HTTP{Addr: "0.0.0.0:8000"},
	})
	if err == nil {
		t.Fatal("expected missing advertiseHost to fail")
	}
}
