package main

import (
	"URLShortener/internal/handlers"
	"URLShortener/internal/repository"
	"URLShortener/internal/service"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Config from env
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0
	redisTTL, _ := time.ParseDuration(os.Getenv("REDIS_TTL"))
	dbPath := os.Getenv("DB_PATH")
	port := os.Getenv("PORT")

	// SQLite
	sqliteRepo, err := repository.NewSQLiteRepo(dbPath)
	if err != nil {
		log.Fatal("Failed to init SQLite:", err)
	}

	redisRepo, err := repository.NewRedisRepo(redisAddr, redisPassword, redisDB, redisTTL)
	if err != nil {
		log.Fatal("Failed to init Redis:", err)
	}

	defer redisRepo.Close()

	svc := service.New(sqliteRepo, redisRepo)

	handler := handlers.New(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("shorten", handler.Shorten)
	mux.HandleFunc("/", handler.Redirect)

	server := http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server stopped gracefully")
}
