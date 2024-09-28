package main

import (
	"context"
	"errors"
	invhandlers "go-http-serve/inventory/handlers"
	invrepositories "go-http-serve/inventory/repositories"
	invusecases "go-http-serve/inventory/usecases"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func main() {
	addr := ":8000"
	readTimeout := 5 * time.Second
	writeTimeout := 5 * time.Second
	shutdownTimeout := 30 * time.Second
	dbConnectTimeout := 30 * time.Second

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	mux := http.NewServeMux()

	dbCtx, dbCancel := context.WithTimeout(context.Background(), dbConnectTimeout)
	defer dbCancel()

	pool, err := pgxpool.New(dbCtx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Error("unable to initialize db pool connection", "err", err)
		return
	}
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	inventoryRepository := invrepositories.NewInventoryRepository(db, log)
	inventoryUseCases := invusecases.NewInventoryCRUDUseCases(inventoryRepository)
	inventoryHandlers := invhandlers.NewInventoryCRUDHandlers(inventoryUseCases, log)
	invhandlers.RegisterRoutes(inventoryHandlers, mux)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}
	go func() {
		log.Info("HTTP Server is listening", "addr", addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("unable to serve HTTP server", "err", err)
			return
		}
		log.Info("HTTP Server stopped")
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	log.Info("Shutting down server...")

	ctx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Shutdown timeout")
		return
	}

	log.Info("Exited, bye.")
}
