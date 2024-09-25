package main

import (
	"context"
	"log"
	"micro-rest-events/v1/app/backend/repository"
	server "micro-rest-events/v1/app/backend/server"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
	_ "github.com/lib/pq"
)

type Server struct {
	PinSize        int
	MaxPinAttempts int
	WebRoot        string
	Version        string
	Listen         string
}

type Options struct {
	Listen         string        `short:"l" long:"listen" env:"LISTEN" default:"localhost:8080" description:"listen address"`
	Secret         string        `short:"s" long:"secret" env:"EVENT_SECRET_KEY" default:"123"`
	PinSize        int           `long:"pinszie" env:"PIN_SIZE" default:"5" description:"pin size"`
	MaxExpire      time.Duration `long:"expire" env:"MAX_EXPIRE" default:"24h" description:"max lifetime"`
	MaxPinAttempts int           `long:"pinattempts" env:"PIN_ATTEMPTS" default:"3" description:"max attempts to enter pin"`
	WebRoot        string        `long:"web" env:"WEB" default:"/" description:"web ui location"`
	Dsn            string        `long:"dsn" env:"POSTGRES_DSN" description:"dsn connection to postgres"`
	Conn           string        `long:"conn" env:"CONNECTION_DSN" default:"micro_events.db" description:"DSN connection, for sqlite use path"`
}

var revision string

func main() {
	log.Printf("Micro rest events %s\n", revision)

	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Fatal(err)
	}

	// if opts.Dsn == "" {
	// 	if err := godotenv.Load(); err != nil {
	// 		panic("No .env file found")
	// 	}
	// } else {
	// 	os.Setenv("POSTGRES_DSN", opts.Dsn)
	// }

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if x := recover(); x != nil {
			log.Printf("[WARN] run time panic:\n%v", x)
			panic(x)
		}

		// catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] interrupt signal")
		cancel()
	}()

	storeProvider, err := repository.NewStoreProvider(opts.Conn)
	if err != nil {
		log.Fatalf("[ERROR] failed to create repository, %+v", err)
	}

	srv := server.Server{
		Listen:         opts.Listen,
		PinSize:        opts.PinSize,
		MaxExpire:      opts.MaxExpire,
		MaxPinAttempts: opts.MaxPinAttempts,
		WebRoot:        opts.WebRoot,
		Secret:         opts.Secret,
		Version:        revision,
		StoreProvider:  storeProvider,
	}
	if err := srv.Run(ctx); err != nil {
		log.Printf("[ERROR] failed, %+v", err)
	}
}
