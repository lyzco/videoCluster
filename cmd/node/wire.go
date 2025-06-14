//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"videoCluster/internal/biz"
	"videoCluster/internal/conf"
	"videoCluster/internal/consul"
	"videoCluster/internal/data"
	"videoCluster/internal/node/server"
	"videoCluster/internal/node/service"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, consul.ProviderSet, newApp))
}
