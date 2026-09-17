package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/signout/signout-api/internal/config"
	authhandler "github.com/signout/signout-api/internal/handler/auth"
	celebhandler "github.com/signout/signout-api/internal/handler/celebration"
	appmw "github.com/signout/signout-api/internal/middleware"
	mongorepo "github.com/signout/signout-api/internal/repository/mongo"
	authsvc "github.com/signout/signout-api/internal/service/auth"
	celebsvc "github.com/signout/signout-api/internal/service/celebration"
	"github.com/signout/signout-api/pkg/response"
)

type Server struct {
	cfg    config.Config
	http   *http.Server
	mongo  *mongorepo.Client
	logger *slog.Logger
}

func New(cfg config.Config, logger *slog.Logger) (*Server, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mongoClient, err := mongorepo.Connect(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		return nil, err
	}
	if err := mongoClient.EnsureIndexes(ctx); err != nil {
		_ = mongoClient.Disconnect(context.Background())
		return nil, err
	}

	tokenMgr := authsvc.NewTokenManager(
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
		cfg.JWTAccessTTL,
		cfg.JWTRefreshTTL,
	)
	users := mongorepo.NewUserRepository(mongoClient.DB)
	refresh := mongorepo.NewRefreshTokenRepository(mongoClient.DB)
	authSvc := authsvc.NewService(users, refresh, tokenMgr)
	authH := authhandler.NewHandler(authSvc)

	shirts := mongorepo.NewShirtRepository(mongoClient.DB)
	if err := shirts.SeedDefaults(ctx); err != nil {
		_ = mongoClient.Disconnect(context.Background())
		return nil, err
	}
	celebrations := mongorepo.NewCelebrationRepository(mongoClient.DB)
	celebSvc := celebsvc.NewService(shirts, celebrations)
	celebH := celebhandler.NewHandler(celebSvc)

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

		api.Get("/shirts", celebH.ListShirts)
		api.Get("/celebrations/{slug}", celebH.GetBySlug)

		api.Group(func(pr chi.Router) {
			pr.Use(appmw.Authenticate(tokenMgr))
			pr.Post("/celebrations", celebH.Create)
			pr.Get("/celebrations", celebH.ListMine)
		})
	})

	httpServer := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      r,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
		IdleTimeout:  cfg.HTTPIdleTimeout,
	}

	return &Server{cfg: cfg, http: httpServer, mongo: mongoClient, logger: logger}, nil
}

func (s *Server) Start() error {
	s.logger.Info("http server listening", slog.String("addr", s.cfg.HTTPAddr))
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	defer func() { _ = s.mongo.Disconnect(ctx) }()
	return s.http.Shutdown(ctx)
}
