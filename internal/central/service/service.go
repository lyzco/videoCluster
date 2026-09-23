package service

import "github.com/google/wire"

// ProviderSet contains application services exposed by the Central service.
var ProviderSet = wire.NewSet(NewVideoService)
