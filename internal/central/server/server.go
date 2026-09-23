package server

import "github.com/google/wire"

// ProviderSet contains transports exposed by the Central service.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer)
