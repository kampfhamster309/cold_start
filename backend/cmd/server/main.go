package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"cold_start/backend/internal/api"
	"cold_start/backend/internal/config"
	"cold_start/backend/internal/db"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "check the running server's health endpoint and exit (for use as a Docker HEALTHCHECK against a shell-less image)")
	flag.Parse()

	cfg := config.Load()

	if *healthcheck {
		os.Exit(runHealthcheck(cfg))
	}

	conn, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer conn.Close()

	log.Print("running database migrations")
	if err := db.Migrate(conn); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	router := api.NewRouter(conn)

	log.Printf("app-backend listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func runHealthcheck(cfg config.Config) int {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + cfg.Port + "/healthz")
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
