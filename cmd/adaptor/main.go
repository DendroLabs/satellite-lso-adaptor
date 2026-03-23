package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DendroLabs/satellite-lso-adaptor/internal/mapping"
	"github.com/DendroLabs/satellite-lso-adaptor/internal/satellite"
	"github.com/DendroLabs/satellite-lso-adaptor/internal/sonata"
	"github.com/DendroLabs/satellite-lso-adaptor/internal/telesat"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := loadConfig()

	// Initialize Telesat VNO API client
	telesatClient := telesat.NewClient(cfg.TelesatBaseURL, cfg.TelesatAPIKey)

	// Initialize satellite context service (Python orbital engine)
	satClient := satellite.NewClient(cfg.SatelliteServiceURL)

	// Initialize the mapping/transformation engine
	transformer := mapping.NewTransformer(telesatClient, satClient)

	// Initialize MEF LSO Sonata API handlers
	mux := http.NewServeMux()
	sonata.RegisterRoutes(mux, transformer)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		slog.Info("shutting down")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	slog.Info("starting satellite LSO adaptor", "port", cfg.Port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

type config struct {
	Port                int
	TelesatBaseURL      string
	TelesatAPIKey       string
	SatelliteServiceURL string
}

func loadConfig() config {
	return config{
		Port:                envInt("PORT", 8080),
		TelesatBaseURL:      envStr("TELESAT_BASE_URL", "https://api.telesat.com/v1"),
		TelesatAPIKey:       envStr("TELESAT_API_KEY", ""),
		SatelliteServiceURL: envStr("SATELLITE_SERVICE_URL", "http://localhost:8090"),
	}
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var i int
		fmt.Sscanf(v, "%d", &i)
		return i
	}
	return fallback
}
