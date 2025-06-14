package cache

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
	"videoCluster/internal/conf"
)

func NewRedisClient(cfg *conf.Data, logger log.Logger) *redis.Client {
	log.Debug("redis config: %v", cfg.Redis)
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       int(cfg.Redis.Db),
	})
}
