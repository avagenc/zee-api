package main

import (
	"context"
	"log"
	"net/http"

	"github.com/avagenc/zee-api/internal/account"
	"github.com/avagenc/zee-api/internal/config"
	"github.com/avagenc/zee-api/internal/device"
	"github.com/avagenc/zee-api/internal/middleware"
	"github.com/avagenc/zee-api/internal/postgres"
	"github.com/avagenc/zee-api/internal/system"
	"github.com/avagenc/zee-api/internal/tuya"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}

	pgPool, err := postgres.NewPool(
		cfg.Database.URL,
		cfg.Database.MaxConns,
		cfg.Database.MinConns,
		cfg.Database.MaxConnLifetime,
		cfg.Database.MaxConnIdleTime,
	)
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}
	defer pgPool.Close()

	if err := postgres.ValidateSchema(context.Background(), pgPool); err != nil {
		log.Fatalf("FATAL: Schema validation failed: %v", err)
	}

	tuyaClient, err := tuya.NewClient(
		cfg.Tuya.AccessID,
		cfg.Tuya.AccessSecret,
		cfg.Tuya.BaseURL,
	)
	if err != nil {
		log.Fatalf("FATAL: Failed to create Tuya client: %v", err)
	}

	accountRepo := account.NewRepository(pgPool)

	svc := struct {
		account account.Service
		device  device.Service
	}{
		account: account.NewService(accountRepo),
		device:  device.NewService(tuyaClient),
	}

	mw := struct {
		tuya      *middleware.Tuya
		requestID func(http.Handler) http.Handler
		realIP    func(http.Handler) http.Handler
		logger    func(http.Handler) http.Handler
		recoverer func(http.Handler) http.Handler
	}{
		tuya:      middleware.NewTuya(accountRepo, tuyaClient),
		requestID: chiMiddleware.RequestID,
		realIP:    chiMiddleware.RealIP,
		logger:    chiMiddleware.Logger,
		recoverer: chiMiddleware.Recoverer,
	}

	hdl := struct {
		system  *system.Handler
		account *account.Handler
		device  *device.Handler
	}{
		system:  system.NewHandler(cfg.App.Name, cfg.App.Version, cfg.App.Env),
		account: account.NewHandler(svc.account),
		device:  device.NewHandler(svc.device),
	}

	r := chi.NewRouter()

	r.Use(mw.requestID)
	r.Use(mw.realIP)
	r.Use(mw.logger)
	r.Use(mw.recoverer)
	r.Use(middleware.AuthenticateAPIKey(cfg.Security.APIKey))

	r.Get("/", hdl.system.Index)

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireUserIdentity)

		r.Get("/account", hdl.account.Get)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.RequireUserIdentity)
		r.Use(mw.tuya.RequireAccount)

		r.Get("/devices", hdl.device.List)

		r.Route("/devices/{deviceId}", func(r chi.Router) {
			r.Use(mw.tuya.RequireDeviceOwnership)
			r.Post("/commands", hdl.device.SendCommands)
		})
	})

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	log.Printf("In the name of Allah, The Most Compassionate, The Most Merciful")
	log.Printf("Starting %s (%s) on port %s\n", cfg.App.Name, cfg.App.Version, cfg.Server.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("FATAL: Failed to start API: %v", err)
	}
}
