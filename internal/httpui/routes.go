package httpui

import (
	"net/http"
	"os"
	"strings"
	"time"

	"ganache-admin-ui/internal/auth"
	"ganache-admin-ui/internal/config"
	"ganache-admin-ui/internal/ganache"
	"ganache-admin-ui/internal/security"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const maxUploadSize = 25 * 1024 * 1024

type Server struct {
	cfg       *config.Config
	users     *auth.UserStore
	sessions  *auth.SessionStore
	client    *ganache.Client
	templates *Templates
}

func NewServer(cfg *config.Config, users *auth.UserStore, sessions *auth.SessionStore, client *ganache.Client) (*Server, error) {
	tmpls, err := ParseTemplates()
	if err != nil {
		return nil, err
	}
	return &Server{cfg: cfg, users: users, sessions: sessions, client: client, templates: tmpls}, nil
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)

	r.Get("/", s.rootRedirect)
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	fs := http.FileServer(http.Dir("web/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	r.Get("/login", s.showLogin)
	r.Post("/login", s.handleLogin)

	r.Group(func(pr chi.Router) {
		pr.Use(auth.RequireAuth(s.sessions, s.path("/login")))
		pr.Use(security.Middleware())

		pr.Post("/logout", s.handleLogout)
		pr.Get("/assets", s.assetsIndex)
		pr.Get("/assets/results", s.assetsResults)
		pr.Get("/assets/new", s.assetsNew)
		pr.Post("/assets/upload", s.assetsUpload)
		pr.Get("/assets/{id}", s.assetDetail)
		pr.Post("/assets/{id}/edit", s.assetEdit)
		pr.Post("/assets/{id}/delete", s.assetDelete)
		pr.Get("/tags", s.tagsList)
	})

	go s.sessionCleanup()

	if s.cfg.BasePath == "" {
		return r
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if trimmed, ok := s.stripBasePath(req.URL.Path); ok {
			clone := req.Clone(req.Context())
			clone.URL.Path = trimmed
			clone.URL.RawPath = trimmed
			if req.URL.RawQuery != "" {
				clone.RequestURI = trimmed + "?" + req.URL.RawQuery
			} else {
				clone.RequestURI = trimmed
			}
			r.ServeHTTP(w, clone)
			return
		}
		r.ServeHTTP(w, req)
	})
}

func (s *Server) rootRedirect(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil {
		if sess, ok := s.sessions.Get(cookie.Value); ok {
			r = r.WithContext(auth.ContextWithSession(r.Context(), sess))
			http.Redirect(w, r, s.path("/assets"), http.StatusFound)
			return
		}
	}
	http.Redirect(w, r, s.path("/login"), http.StatusFound)
}

func (s *Server) sessionCleanup() {
	ticker := time.NewTicker(30 * time.Minute)
	for range ticker.C {
		s.sessions.CleanupExpired()
	}
}

func (s *Server) stripBasePath(p string) (string, bool) {
	base := s.cfg.BasePath
	if base == "" {
		return "", false
	}
	if p == base {
		return "/", true
	}
	if strings.HasPrefix(p, base+"/") {
		trimmed := strings.TrimPrefix(p, base)
		if trimmed == "" {
			trimmed = "/"
		}
		return trimmed, true
	}
	return "", false
}

func (s *Server) path(rel string) string {
	if rel == "" {
		rel = "/"
	}
	if !strings.HasPrefix(rel, "/") {
		rel = "/" + rel
	}
	base := s.cfg.BasePath
	if base == "" {
		return rel
	}
	if rel == "/" {
		return base
	}
	return base + rel
}

func (s *Server) cookiePath() string {
	if s.cfg.BasePath == "" {
		return "/"
	}
	return s.cfg.BasePath
}

func (s *Server) render(w http.ResponseWriter, name string, data TemplateData, r *http.Request) {
	data.BasePath = s.cfg.BasePath
	s.templates.Render(w, name, data, r)
}

func secureCookie() bool {
	return os.Getenv("UI_SECURE_COOKIE") == "true"
}
