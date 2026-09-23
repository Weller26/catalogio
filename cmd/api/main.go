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
		7*24*time.Hour,
	)

	authHandler := auth.NewHandler(
		authService,
		false,
	)

	mux := http.NewServeMux()

	mux.HandleFunc(
		"POST /api/v1/auth/register",
		authHandler.Register,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/login",
		authHandler.Login,
	)

	mux.HandleFunc(
		"POST /api/v1/auth/logout",
		authHandler.Logout,
	)

	meHandler := http.HandlerFunc(
		authHandler.Me,
	)

	mux.Handle(
		"GET /api/v1/me",
		authService.RequireAuth(meHandler),
	)

	server := http.Server{
		Addr: fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")),
		Handler: mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("server started on :%s", os.Getenv("SERVER_PORT"))

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}