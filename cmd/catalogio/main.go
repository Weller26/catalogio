package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Weller26/catalogio/internal/auth"
	"github.com/Weller26/catalogio/internal/catalog"
	"github.com/Weller26/catalogio/internal/database"
	"github.com/Weller26/catalogio/internal/httpapi"
	"github.com/Weller26/catalogio/internal/metrics"
)

func main() {
	var level slog.Level
	switch strings.ToUpper(os.Getenv("LOG_LEVEL")) {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				a.Key = "timestamp"
			case slog.MessageKey:
				a.Key = "message"
			}
			return a
		},
	})

	logger := slog.New(handler).With(
		slog.String("service.name", "catalogio"),
	)

	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	cfg := database.Config {
		Host: os.Getenv("DB_HOST"),
		Port: os.Getenv("DB_PORT"),
		User: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName: os.Getenv("DB_NAME"),
		SSLMode: os.Getenv("DB_SSL_MODE"),
		MaxConns: 10,
		MinConns: 2,
		MaxConnIdleTime: 5 * time.Minute,
		MaxConnLifetime: 1 * time.Hour,
	}

	db, err := database.New(ctx, cfg)
	if err != nil {
		slog.ErrorContext(ctx, "failed to connect to the database",
			"error", err,
		)
		os.Exit(1)
	}
	defer db.Close()

	slog.InfoContext(ctx, "successful connection to PostgreSQL")

	initEntityCount := func(table, label string) {
		var count float64
		query := fmt.Sprintf("SELECT count(*) FROM %s", table)
		if err := db.Pool.QueryRow(ctx, query).Scan(&count); err == nil {
			metrics.EntitiesCurrent.WithLabelValues(label).Set(count)
		}
	}

	initEntityCount("items", "item")
	initEntityCount("item_statuses", "item-status")
	initEntityCount("item_types", "item-type")

	authRepo := auth.NewRepository(db.Pool)
	authService := auth.NewService(
		authRepo,
		os.Getenv("JWT_SECRET"),
	)
	authHandler := auth.NewHandler(
		authService,
		false,
	)

	itemRepository := catalog.NewRepository(db.Pool)
	itemService := catalog.NewService(itemRepository)
	itemHandler := catalog.NewHandler(itemService)

	authMiddleware := httpapi.NewAuthMiddleware(authService)

	router := httpapi.NewRouter(
		authHandler,
		itemHandler,
		authMiddleware,
	)

	server := http.Server{
		Addr: fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")),
		Handler: router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.InfoContext(ctx, "server started",
			"server_port", os.Getenv("SERVER_PORT"))

		if err := server.ListenAndServe(); err != nil && 
			err != http.ErrServerClosed {
			slog.ErrorContext(ctx, "server error",
				"error", err,
			)
		}
	}()

	<-ctx.Done()

	slog.InfoContext(ctx, "shutting down server")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10 * time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.ErrorContext(ctx, "server shutdown error",
			"error", err,
		)
	}

	slog.InfoContext(ctx, "server stopped")
}