package main

import (
	"context"
	"flag"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	_ "go.uber.org/automaxprocs"
	"os"
	"videoCluster/internal/consul"
	"videoCluster/internal/node/biz"
	"videoCluster/internal/node/conf"
	"videoCluster/internal/node/job"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "config.yaml", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, registrar *consul.NodeRegistrar, vs *biz.VideoScanner, data *conf.Data) *kratos.App {
	opts := []kratos.Option{
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
		kratos.Registrar(registrar),
	}

	if data.ScanVideo {
		if data.ScanInterval == nil || data.ScanInterval.AsDuration() <= 0 {
			panic("data.scanInterval must be a positive duration when data.scanVideo is true")
		}

		jobManager := job.NewJobManager(func() {
			if err := vs.ScanAndCache(context.Background()); err != nil {
				log.Context(context.Background()).Errorf("scan videos failed: %v", err)
			}
		}, data.ScanInterval.AsDuration(), false)
		opts = append(opts,
			kratos.BeforeStart(func(ctx context.Context) error {
				jobManager.Start()
				return nil
			}),
			kratos.AfterStop(func(ctx context.Context) error {
				jobManager.Stop()
				return nil
			}),
		)
	}

	return kratos.New(opts...)
}

func main() {
	flag.Parse()

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	Name = bc.Server.ServerName
	id = bc.Server.ServerId

	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
