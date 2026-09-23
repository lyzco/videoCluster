package consul

import (
	"github.com/google/wire"
)

// ProviderSet is server providers.
var CentralProviderSet = wire.NewSet(NewCentralConsulClient)

// NodeProviderSet contains the Consul client and Node-specific registrar.
var NodeProviderSet = wire.NewSet(NewNodeConsulClient, NewNodeRegistrar)
