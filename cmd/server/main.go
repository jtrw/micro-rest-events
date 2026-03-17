package main

import (
	"context"
	"log/slog"
	"micro-rest-events/internal/logger"
	"micro-rest-events/internal/repository"
	"micro-rest-events/internal/web"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"
	_ "github.com/lib/pq"
)

type Options struct {
	Listen string `short:"l" long:"listen" env:"LISTEN" default:":8181" description:"listen address"`
	Secret string `short:"s" long:"secret" env:"EVENT_SECRET_KEY" default:"123"`
	Debug  bool   `long:"dbg" env:"DEBUG" description:"enable debug logging"`
	Dsn    string `long:"dsn" env:"POSTGRES_DSN" description:"dsn connection to postgres"`
	Conn   string `long:"conn" env:"CONNECTION_DSN" default:"micro_events.db" description:"DSN connection, for sqlite use path"`
}

var revision string

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	var h slog.Handler
	if opts.Debug {
		h = logger.NewDebugHandler(os.Stdout)
	} else {
		h = logger.NewHandler(os.Stdout, slog.LevelInfo)
	}
	slog.SetDefault(slog.New(h))

	slog.Info("starting", "revision", revision, "listen", opts.Listen)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if x := recover(); x != nil {
			slog.Warn("run time panic", "panic", x)
			panic(x)
		}
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		slog.Warn("interrupt signal")
		cancel()
	}()

	storeProvider, err := repository.NewStoreProvider(opts.Conn)
	if err != nil {
		slog.Error("failed to create repository", "err", err)
		os.Exit(1)
	}

	srv := &web.Server{
		Listen:        opts.Listen,
		Secret:        opts.Secret,
		Version:       revision,
		StoreProvider: storeProvider,
	}
	if err := srv.Run(ctx); err != nil {
		slog.Error("server failed", "err", err)
	}
}
