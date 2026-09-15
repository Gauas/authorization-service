package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gauas/authorization-service/forwardauth"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("authorization service stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := forwardauth.ConfigFromEnv()
	if err != nil {
		return err
	}
	redisOptions, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return err
	}
	redisOptions.DisableIdentity = true
	client := redis.NewClient(redisOptions)
	defer client.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	registry := prometheus.NewRegistry()
	metrics := forwardauth.NewMetrics(registry)
	keys := forwardauth.NewJWKSCache(cfg.JWKSURL, cfg.JWKSRequestTimeout)
	keys.SetMetrics(metrics)
	if err := keys.Refresh(ctx); err != nil {
		slog.Error("initial JWKS refresh failed", "error", err)
	}
	go keys.Run(ctx, cfg.JWKSRefresh)
	consumer, err := forwardauth.NewRevocationConsumer(cfg.QueueURL, cfg.SecurityQueue, cfg.KafkaConsumerGroup)
	if err != nil {
		return err
	}
	defer consumer.Close()
	events := forwardauth.NewRevocationEventHandler(client, cfg.RevocationPrefix, cfg.RedisLookupTimeout)
	go consumeSecurityEvents(ctx, stop, consumer, events)
	revocations := forwardauth.NewRedisRevocations(client, cfg.RevocationPrefix, cfg.RedisLookupTimeout)
	handler := forwardauth.NewHandler(
		forwardauth.NewVerifier(keys, cfg),
		revocations.Revoked,
		metrics,
	)

	mux := http.NewServeMux()
	mux.Handle("GET /v1/authorization/forward-auth", handler)
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })

	server := &http.Server{
		Addr: cfg.Address, Handler: forwardauth.HTTPMiddleware(mux), ReadHeaderTimeout: cfg.ReadHeaderTimeout, IdleTimeout: cfg.IdleTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	stopped := make(chan struct{})
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		_ = server.Shutdown(shutdown)
		close(stopped)
	}()

	slog.Info("authorization service listening", "address", cfg.Address)
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	<-stopped
	return nil
}

func consumeSecurityEvents(ctx context.Context, stop context.CancelFunc, consumer forwardauth.RevocationConsumer, events *forwardauth.RevocationEventHandler) {
	err := consumer.Run(ctx, events.Handle)
	if err == nil || ctx.Err() != nil {
		return
	}
	slog.Error("security event consumer stopped", "error", err)
	stop()
}
