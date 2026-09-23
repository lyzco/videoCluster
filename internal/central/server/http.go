package server

import (
	"videoCluster/internal/central/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer creates the Central HTTP transport. HTTP endpoints will be
// registered when the Central API gains HTTP annotations.
func NewHTTPServer(c *conf.Server, logger log.Logger) *http.Server {
	opts := []http.ServerOption{
		http.Middleware(recovery.Recovery()),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	return http.NewServer(opts...)
}
