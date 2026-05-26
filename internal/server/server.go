package server

import (
	"embed"
	"io/fs"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/burak/linux-dashboard/internal/ai"
	"github.com/burak/linux-dashboard/internal/collector"
	"github.com/burak/linux-dashboard/internal/event"
	"github.com/burak/linux-dashboard/internal/storage"
)

//go:embed all:web
var webAssets embed.FS

// Server is the HTTP server hosting the REST API and SSE stream.
type Server struct {
	addr      string
	router    *chi.Mux
	collector *collector.Manager
	store     *storage.Store
	aiAdvisor *ai.Advisor
	emitter   *event.Emitter
}

// New creates a new Server with all dependencies wired up.
func New(addr string, col *collector.Manager, store *storage.Store, advisor *ai.Advisor, emitter *event.Emitter) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Compress(5))

	WireRoutes(r, col, store, advisor, emitter)

	// Serve embedded frontend static files for non-API routes.
	// API routes (registered before this) take priority.
	subFS, err := fs.Sub(webAssets, "web")
	if err == nil {
		fs := http.FileServer(http.FS(subFS))
		r.Group(func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if len(r.URL.Path) >= 4 && r.URL.Path[:4] != "/api" {
						fs.ServeHTTP(w, r)
						return
					}
					next.ServeHTTP(w, r)
				})
			})
			r.Handle("/*", http.NotFoundHandler())
		})
	}

	return &Server{
		addr:      addr,
		router:    r,
		collector: col,
		store:     store,
		aiAdvisor: advisor,
		emitter:   emitter,
	}
}

// Listen starts the HTTP server and blocks until it exits.
func (s *Server) Listen() error {
	srv := &http.Server{
		Addr:         s.addr,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return srv.ListenAndServe()
}