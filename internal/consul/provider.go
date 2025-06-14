package consul

import (
	"github.com/hashicorp/consul/api"
	"log"
	"videoCluster/internal/conf"
)

func NewConsulClient(cfg *conf.Data) (*api.Client, error) {
	consulCfg := api.DefaultConfig()
	consulCfg.Address = cfg.Consul.Address
	consulCfg.Token = cfg.Consul.Token

	client, err := api.NewClient(consulCfg)
	if err != nil {
		return nil, err
	}
	log.Printf("Consul client initialized at %s", cfg.Consul.Address)
	return client, nil
}
