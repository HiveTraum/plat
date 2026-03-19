package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"8080"`
}

type App struct {
	DB *pgxpool.Pool
}

func NewApp() (*App, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://plat:plat@localhost:5432/plat?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &App{DB: pool}, nil
}

type HealthOutput struct {
	Body struct {
		Status string `json:"status" example:"ok" doc:"Service health status"`
	}
}

type UserOutput struct {
	Body struct {
		ID    string `json:"id" doc:"User ID"`
		Email string `json:"email" doc:"User email"`
	}
}

func main() {
	cli := humacli.New(func(hooks humacli.Hooks, options *Options) {
		app, err := NewApp()
		if err != nil {
			log.Fatalf("Failed to initialize app: %v", err)
		}

		router := http.NewServeMux()
		api := humachi.New(router, huma.DefaultConfig("Plat API", "1.0.0"))

		huma.Get(api, "/api/health", func(ctx context.Context, input *struct{}) (*HealthOutput, error) {
			err := app.DB.Ping(ctx)
			if err != nil {
				return nil, huma.Error503ServiceUnavailable("database is unavailable")
			}
			resp := &HealthOutput{}
			resp.Body.Status = "ok"
			return resp, nil
		})

		huma.Get(api, "/api/me", func(ctx context.Context, input *struct {
			UserID    string `header:"X-User-Id" required:"true" doc:"User ID from Oathkeeper"`
			UserEmail string `header:"X-User-Email" required:"true" doc:"User email from Oathkeeper"`
		}) (*UserOutput, error) {
			resp := &UserOutput{}
			resp.Body.ID = input.UserID
			resp.Body.Email = input.UserEmail
			return resp, nil
		})

		hooks.OnStart(func() {
			addr := fmt.Sprintf(":%d", options.Port)
			log.Printf("Starting server on %s", addr)
			if err := http.ListenAndServe(addr, router); err != nil {
				log.Fatalf("Server failed: %v", err)
			}
		})

		hooks.OnStop(func() {
			app.DB.Close()
		})
	})

	cli.Run()
}
