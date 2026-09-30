package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Weller26/catalogio/internal/auth"
	"github.com/Weller26/catalogio/internal/database"
	"github.com/Weller26/catalogio/internal/httpapi"
	"github.com/Weller26/catalogio/internal/items"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := database.Config {
		Host: os.Getenv("DB_HOST"),
		Port: os.Getenv("DB_PORT"),
		User: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName: os.Getenv("DB_NAME"),
		SSLMode: "disable",
		MaxConns: 10,
		MinConns: 2,
		MaxConnIdleTime: 5 * time.Minute,
		MaxConnLifetime: 1 * time.Hour,
	}

	db, err := database.New(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer db.Close()

	log.Println("Successful connection to PostgreSQL")

	authRepo := auth.NewRepository(db.Pool)
	authService := auth.NewService(
		authRepo,
		os.Getenv("JWT_SECRET"),
	)
	authHandler := auth.NewHandler(
		authService,
		false,
	)

	itemRepository := items.NewRepository(db.Pool)
	itemService := items.NewService(itemRepository)
	itemHandler := items.NewHandler(itemService)

	router := httpapi.NewRouter(
		authService,
		authHandler,
		itemHandler,
	)

	server := http.Server{
		Addr: fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")),
		Handler: router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf(
			"server started on: %s",
			os.Getenv("SERVER_PORT"),
		)

		if err := server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatalf(
				"server error: %v",
				err,
			)
		}
	}()

	<-ctx.Done()

	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10 * time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf(
			"server shutdown error: %v",
			err,
		)
	}

	log.Println("server stopped")
}