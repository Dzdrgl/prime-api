package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func normalizeAddr(port string) string {
	if port == "" {
		return ":8080"
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}

func NewHTTPServer(handler http.Handler, port string) *http.Server {
	return &http.Server{
		Addr:    normalizeAddr(port),
		Handler: handler,
	}
}

func RunHTTPServer(logger *zap.Logger, srv *http.Server, wait time.Duration) {
	logger.Info("server starting", zap.String("addr", srv.Addr))

	// dinleme
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Info("shutdown signal received", zap.String("signal", sig.String()))
	case err := <-errCh:
		logger.Fatal("listen error", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	} else {
		logger.Info("server exited")
	}
}
