package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"log"
	centralconf "videoCluster/internal/central/conf"
	nodeconf "videoCluster/internal/node/conf"
)

func NewCentralConsulClient(cfg *centralconf.Data) (*api.Client, error) {
	if cfg == nil || cfg.Consul == nil {
		return nil, fmt.Errorf("data.consul.address is required")
	}
	return newConsulClient(cfg.Consul.Address, cfg.Consul.Token)
}

func NewNodeConsulClient(cfg *nodeconf.Data) (*api.Client, error) {
	if cfg == nil || cfg.Consul == nil {
		return nil, fmt.Errorf("data.consul.address is required")
	}
	return newConsulClient(cfg.Consul.Address, cfg.Consul.Token)
}

func newConsulClient(address, token string) (*api.Client, error) {
	if address == "" {
		return nil, fmt.Errorf("data.consul.address is required")
	}
	consulCfg := api.DefaultConfig()
	consulCfg.Address = address
	consulCfg.Token = token

	client, err := api.NewClient(consulCfg)
	if err != nil {
		return nil, err
	}
	log.Printf("Consul client initialized at %s", address)
	return client, nil
}
