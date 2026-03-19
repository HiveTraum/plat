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
	ory "github.com/ory/client-go"
)

type Options struct {
	Port int `help:"Port to listen on" short:"p" default:"8080"`
}

type App struct {
	DB     *pgxpool.Pool
	Kratos *ory.APIClient
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

	kratosURL := os.Getenv("KRATOS_PUBLIC_URL")
	if kratosURL == "" {
		kratosURL = "http://localhost:4433"
	}

	kratosConfig := ory.NewConfiguration()
	kratosConfig.Servers = ory.ServerConfigurations{{URL: kratosURL}}
	kratosClient := ory.NewAPIClient(kratosConfig)

	return &App{
		DB:     pool,
		Kratos: kratosClient,
	}, nil
}

type HealthOutput struct {
	Body struct {
		Status string `json:"status" example:"ok" doc:"Service health status"`
	}
}

type UserOutput struct {
	Body struct {
		ID    string `json:"id" doc:"User ID from Kratos"`
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
			Cookie string `header:"Cookie" doc:"Session cookie"`
		}) (*UserOutput, error) {
			session, _, err := app.Kratos.FrontendAPI.ToSession(ctx).Cookie(input.Cookie).Execute()
			if err != nil {
				return nil, huma.Error401Unauthorized("unauthorized")
			}

			identity := session.Identity
			resp := &UserOutput{}
			resp.Body.ID = identity.Id

			if traits, ok := identity.Traits.(map[string]interface{}); ok {
				if email, ok := traits["email"].(string); ok {
					resp.Body.Email = email
				}
			}

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
