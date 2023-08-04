package server

import (
   "context"
   "time"
   "log"
   "net/http"
   "github.com/pkg/errors"
   "github.com/didip/tollbooth/v7"
   "github.com/didip/tollbooth_chi"
   "github.com/go-chi/chi/v5"
   "github.com/go-chi/chi/v5/middleware"
   "github.com/go-chi/render"
   "github.com/jtrw/go-rest"
   repository "go-rest-events/v1/app/backend/repository"
   event_handler "go-rest-events/v1/app/backend/handler"
)

type Server struct {
    Port           string
	PinSize        int
	MaxPinAttempts int
	MaxExpire      time.Duration
	WebRoot        string
	Version        string
	Repository     repository.Repository

}

func (s Server) Run(ctx context.Context) error {
	log.Printf("[INFO] activate rest server")
    log.Printf("[INFO] Port: %s", s.Port)

	httpServer := &http.Server{
		Addr:              ":"+s.Port,
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
	return nil
}

func (s Server) routes() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.RequestID, middleware.RealIP)
	router.Use(middleware.Throttle(1000), middleware.Timeout(60*time.Second))
	//router.Use(rest.AppInfo("events", "Jrtw", s.Version), rest.Ping)
	router.Use(rest.Ping)
	router.Use(tollbooth_chi.LimitHandler(tollbooth.NewLimiter(10, nil)))
    router.Use(middleware.Logger)
    handler := event_handler.NewHandler(s.Repository.Connection)
	router.Route("/api/v1", func(r chi.Router) {
	    //r.Use(Authentication)
	    r.Use(Cors)
        r.Post("/events", handler.OnCreateEvent)
        r.Post("/events/{uuid}", handler.OnChangeEvent)
        r.Get("/events/users/{id}", handler.OnGetEventsByUserId)
	})

	router.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, "User-agent: *\nDisallow: /api/\nDisallow: /show/\n")
	})

	return router
}
