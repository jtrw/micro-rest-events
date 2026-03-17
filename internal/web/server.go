package web

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	provider "micro-rest-events/internal/repository"
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

//go:embed templates static
var embedFS embed.FS

type Server struct {
	Listen        string
	Secret        string
	Version       string
	AuthLogin     string
	AuthPassword  string
	StoreProvider provider.StoreProviderInterface
	tmpl          *template.Template
}

func (s *Server) Run(ctx context.Context) error {
	slog.Info("activate rest server", "listen", s.Listen)

	tmpl, err := template.New("").ParseFS(embedFS, "templates/*.html", "templates/partials/*.html")
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}
	s.tmpl = tmpl

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
				slog.Error("failed to close http server", "err", clsErr)
			}
		}
	}()

	err = httpServer.ListenAndServe()
	slog.Warn("http server terminated", "err", err)

	if err != http.ErrServerClosed {
		return errors.Wrap(err, "server failed")
	}
	return err
}

func (s *Server) routes() chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.RequestID, middleware.RealIP)
	router.Use(middleware.Throttle(1000), middleware.Timeout(60*time.Second))
	router.Use(rest.Ping)
	router.Use(tollbooth_chi.LimitHandler(tollbooth.NewLimiter(10, nil)))
	router.Use(middleware.Logger)
	router.Use(Cors)

	// Static files
	router.Handle("/static/*", http.FileServerFS(embedFS))

	// Auth (form-based)
	router.Get("/login", s.handleLoginPage)
	router.Post("/login", s.handleLoginSubmit)
	router.Get("/logout", s.handleLogout)

	// Web UI (protected by session cookie via AuthMiddleware)
	router.Group(func(r chi.Router) {
		r.Use(s.AuthMiddleware)
		r.Get("/", s.dashboard)
		r.Get("/web/events", s.eventsTable)
		r.Post("/web/events", s.createEvent)
		r.Post("/web/events/{uuid}/status", s.changeStatus)
		r.Post("/web/events/{uuid}/seen", s.markSeen)
	})

	// API
	h := Handler{StoreProvider: s.StoreProvider}
	router.Route("/api/v1", func(r chi.Router) {
		r.Use(rest.AuthenticationJwt("Api-Token", s.Secret, func(claims map[string]interface{}) error {
			if claims["user_id"] == nil {
				return fmt.Errorf("user_id not found")
			}
			return nil
		}))

		r.Post("/events", h.OnCreateEvent)
		r.Post("/events/batch", h.OnCreateBatchEvents)
		r.Post("/events/{uuid}", h.OnChangeEvent)
		r.Get("/events/users/{id}", h.OnGetEventsByUserId)
		r.Post("/events/{uuid}/seen", h.OnSetSeen)
		r.Post("/events/change/batch", h.OnChangeBatchEvents)
	})

	router.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, "User-agent: *\nDisallow: /\n")
	})

	return router
}
