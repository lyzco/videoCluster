package data

import (
	"github.com/dgraph-io/badger/v4"
	"videoCluster/internal/conf"
)

func NewBadgerDB(c *conf.Server) (*badger.DB, error) {
	opts := badger.DefaultOptions(c.ServerName + c.ServerId + "video.data").WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return db, nil
}
