package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/signout/signout-api/internal/config"
	authhandler "github.com/signout/signout-api/internal/handler/auth"
	appmw "github.com/signout/signout-api/internal/middleware"
	"github.com/signout/signout-api/internal/repository/postgres"
	authsvc "github.com/signout/signout-api/internal/service/auth"
	"github.com/signout/signout-api/pkg/response"
)

type Server struct {
	cfg    config.Config
	http   *http.Server
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func New(cfg config.Config, logger *slog.Logger) (*Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if err := postgres.RunMigrations(ctx, pool, "migrations"); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrations: %w", err)
	}

	tokenMgr := authsvc.NewTokenManager(
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
		cfg.JWTAccessTTL,
		cfg.JWTRefreshTTL,
	)
	users := postgres.NewUserRepository(pool)
	refresh := postgres.NewRefreshTokenRepository(pool)
	svc := authsvc.NewService(users, refresh, tokenMgr)
	authH := authhandler.NewHandler(svc)

	r := chi.NewRouter()
	r.Use(appmw.Recover)
	r.Use(appmw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/v1", func(api chi.Router) {
		api.Route("/auth", func(ar chi.Router) {
			ar.Post("/register", authH.Register)
			ar.Post("/login", authH.Login)
			ar.Post("/refresh", authH.Refresh)
			ar.Post("/logout", authH.Logout)

			ar.Group(func(pr chi.Router) {
				pr.Use(appmw.Authenticate(tokenMgr))
				pr.Get("/me", authH.Me)
			})
		})
	})

	httpServer := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
		IdleTimeout:  cfg.HTTPIdleTimeout,
	}

	return &Server{cfg: cfg, http: httpServer, pool: pool, logger: logger}, nil
}

func (s *Server) Start() error {
	s.logger.Info("http server listening", slog.String("addr", s.cfg.HTTPAddr))
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	defer s.pool.Close()
	return s.http.Shutdown(ctx)
}
