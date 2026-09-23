//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"videoCluster/internal/central/biz"
	"videoCluster/internal/central/conf"
	"videoCluster/internal/central/data"
	"videoCluster/internal/central/server"
	"videoCluster/internal/central/service"
	"videoCluster/internal/consul"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		consul.CentralProviderSet,
		wire.Bind(new(biz.NodeSource), new(*data.NodeRepository)),
		newApp,
	))
}
