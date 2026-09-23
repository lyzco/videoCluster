package data

import (
	"github.com/google/wire"
	"videoCluster/internal/node/biz"
)

// ProviderSet contains persistence providers used by the Node service.
var ProviderSet = wire.NewSet(
	NewBadgerDB,
	NewVideoVideoRepository,
	wire.Bind(new(biz.VideoRepository), new(*VideoVideoRepository)),
)
