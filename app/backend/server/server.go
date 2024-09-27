package server

import (
	"context"
	"log"
	event_handler "micro-rest-events/v1/app/backend/handler"
	repository "micro-rest-events/v1/app/backend/repository"
	"net/http"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth_chi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/jtrw/go-rest"
	"github.com/pkg/errors"
)

type Server struct {
	Listen         string
	PinSize        int
	MaxPinAttempts int
	MaxExpire      time.Duration
	WebRoot        string
	Secret         string
	Version        string
	StoreProvider  repository.StoreProviderInterface
}

func (s Server) Run(ctx context.Context) error {
	log.Printf("[INFO] activate rest server")
	log.Printf("[INFO] Listen: %s", s.Listen)

	httpServer := &http.Server{
		Addr:              s.Listen,
		Handler:           s.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	go func() {
		<-ctx.Done()
		if httpServer != nil {
			if clsErr := httpServer.Close(); clsErr != nil {
				log.Printf("[ERROR] failed to close proxy http server, %v", clsErr)
			}
		}
	}()

	err := httpServer.ListenAndServe()
	log.Printf("[WARN] http server terminated, %s", err)

	if err != http.ErrServerClosed {
		return errors.Wrap(err, "server failed")
	}
	return err
}

func (s Server) routes() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.RequestID, middleware.RealIP)
	router.Use(middleware.Throttle(1000), middleware.Timeout(60*time.Second))
	//router.Use(rest.AppInfo("events", "Jrtw", s.Version), rest.Ping)
	router.Use(rest.Ping)
	router.Use(tollbooth_chi.LimitHandler(tollbooth.NewLimiter(10, nil)))
	router.Use(middleware.Logger)
	//er := event.NewEventRepository(s.StoreProvider)
	handler := event_handler.NewHandler(s.StoreProvider)
	router.Route("/api/v1", func(r chi.Router) {
		// r.Use(rest.AuthenticationJwt("Api-Token", s.Secret, func(claims map[string]interface{}) error {
		// 	if claims["user_id"] == nil {
		// 		return fmt.Errorf("user_id not found")
		// 	}
		// 	return nil
		// }))
		r.Use(Cors)
		r.Post("/events", handler.OnCreateEvent)
		r.Post("/events/batch", handler.OnCreateBatchEvents)
		r.Post("/events/{uuid}", handler.OnChangeEvent)
		r.Get("/events/users/{id}", handler.OnGetEventsByUserId)
		r.Post("/events/{uuid}/seen", handler.OnSetSeen)
		r.Post("/events/change/batch", handler.OnChangeBatchEvents)
	})

	router.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, "User-agent: *\nDisallow: /\n")
	})

	return router
}
