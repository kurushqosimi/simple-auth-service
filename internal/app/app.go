package app

import (
	"context"
	"errors"
	"fmt"
	"fullstack-simple-app/config"
	"fullstack-simple-app/internal/adapters"
	"fullstack-simple-app/internal/repositories"
	"fullstack-simple-app/internal/services"
	"fullstack-simple-app/internal/transport/http"
	"fullstack-simple-app/pkg/async"
	"fullstack-simple-app/pkg/email"
	"fullstack-simple-app/pkg/logger"
	"fullstack-simple-app/pkg/postgres"
	"fullstack-simple-app/pkg/redis"
	"fullstack-simple-app/pkg/tokens/authentication"
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"
	"net"
	http3 "net/http"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
)

type App struct {
	cfg        *config.Config
	wg         sync.WaitGroup
	router     *gin.Engine
	httpServer *http3.Server
	logger     logger.Logger
	pg         *postgres.Postgres
	redis      *redis.RedisClient
}

func New(cfg *config.Config) (*App, error) {
	var a = &App{
		wg: sync.WaitGroup{},
	}

	l := logger.New(logger.WithLevel(cfg.Log.Level), logger.WithIsJSON(true), logger.WithAddSource(true))
	l.Debug("logger initialized")

	l.Debug("Configurations: %v", cfg)

	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		l.Fatal(fmt.Sprintf("app - Run - postgres.New: %v", err))
	}

	redisClient := redis.New(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)

	mailer := email.NewMailer(cfg.Mailer.SMTPServer, cfg.Mailer.SMTPPort, cfg.Mailer.SenderMail, cfg.Mailer.SenderPassword, cfg.Mailer.SenderName)

	runner := async.NewBackgroundRunner(&a.wg)

	tokenMaker, err := authentication.NewPasetoMaker(cfg.TokenKey.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	userRepo := repositories.NewUserRepo(pg)
	emailSender := adapters.NewEmailAdapter(mailer)
	userService := services.NewUserService(userRepo, emailSender, runner, redisClient, tokenMaker)
	userHandler := http.NewUserHandler(userService, l)

	router := http.NewRouter(userHandler)

	a.cfg = cfg
	a.router = router
	a.logger = l
	a.pg = pg
	a.redis = redisClient

	return a, nil
}

func (a *App) Run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	grp, ctx := errgroup.WithContext(ctx)

	grp.Go(func() error {
		return a.startHTTP(ctx)
	})

	err := grp.Wait()
	switch {
	case err == nil || errors.Is(err, context.Canceled):
		a.logger.Info("Everything is good. Server shutting down...")
	case errors.Is(err, http3.ErrServerClosed):
		a.logger.Info("server shutdown")
	default:
		a.logger.Error(fmt.Sprintf("app.Run: %v", err))
		return err
	}

	a.logger.Info("waiting for background tasks to finish...")
	a.wg.Wait()

	if a.pg != nil {
		a.pg.Close()
		a.logger.Info("Postgres connection closed")
	}

	if a.redis != nil {
		err = a.redis.Close()
		if err != nil {
			a.logger.Info("Redis connection could not close: ", err)
		}
		a.logger.Info("Redis connection closed")
	}

	return nil
}

func (a *App) startHTTP(ctx context.Context) error {
	const op = "startHTTP"

	a.logger.Info("HTTP Server initializing")

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", "0.0.0.0", a.cfg.HTTP.Port))
	if err != nil {
		a.logger.Fatal("%s: failed to create listener: ", op, err)
		return err
	}

	c := cors.New(cors.Options{
		AllowedMethods:     a.cfg.HTTP.CORS.AllowedMethods,
		AllowedOrigins:     a.cfg.HTTP.CORS.AllowedOrigins,
		AllowCredentials:   a.cfg.HTTP.CORS.AllowCredentials,
		AllowedHeaders:     a.cfg.HTTP.CORS.AllowedHeaders,
		OptionsPassthrough: a.cfg.HTTP.CORS.OptionsPassthrough,
		ExposedHeaders:     a.cfg.HTTP.CORS.ExposedHeaders,
		Debug:              a.cfg.HTTP.CORS.Debug,
	})

	handler := c.Handler(a.router)

	a.httpServer = &http3.Server{
		Handler:      handler,
		WriteTimeout: a.cfg.HTTP.WriteTimeout,
		ReadTimeout:  a.cfg.HTTP.ReadTimeout,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		a.logger.Info("Shutting down HTTP server...")
		err = a.httpServer.Shutdown(shutdownCtx)
		if err != nil {
			a.logger.Error("failed to shutdown server: %v", err)
		}
	}()

	if err = a.httpServer.Serve(listener); err != nil {
		switch {
		case err == nil:
			a.logger.Info("server exited with no error")
			return nil
		case errors.Is(err, http3.ErrServerClosed):
			a.logger.Info("server shutdown")
		default:
			a.logger.Fatal("failed to start server")
		}
	}

	return err
}

func (a *App) background(fn func()) {
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()

		defer func() {
			if err := recover(); err != nil {
				a.logger.Error(fmt.Sprintf("panic in background task: %v\n%s", err, debug.Stack()))
			}
		}()
		fn()
	}()
}
