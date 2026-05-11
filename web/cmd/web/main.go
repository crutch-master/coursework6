package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/crutch-master/coursework6/web/internal/wiring"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)
	defer stop()

	handler, err := wiring.Wire(ctx)
	if err != nil {
		slog.Error("failed to initalize", "err", fmt.Errorf("wiring.Wire: %w", err))
		return
	}

	server := http.Server{
		Addr:    ":9090",
		Handler: handler,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("failed to shut down the server", "err", err)
		}
	}()

	slog.Info("starting the app")

	err = server.ListenAndServe()
	if err != nil {
		slog.Error("server stopped", "err", fmt.Errorf("server.ListenAndServe: %w", err))
	}
}
