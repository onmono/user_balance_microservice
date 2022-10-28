package main

import (
	"context"
	"fmt"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/onmono/internal/balance"
	"github.com/onmono/internal/balance/db"
	"github.com/onmono/internal/routes"
	"github.com/onmono/internal/usecases"
	"github.com/onmono/pkg/client/database/postgresql"
	"github.com/onmono/pkg/logging"
	"log"
	"net/http"
	"os"
)

const webPort = "80"

func main() {
	log.Println("Starting user-balance-microservice...")
	logger := logging.GetLogger()
	ctx := context.Background()

	repository := connectToDB(ctx, &logger)
	if repository == nil {
		log.Panic("Can't connect to Postgres!")
	}

	uc := usecases.NewUseCase(ctx, repository, &logger)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: routes.Routes(uc, &logger),
	}

	log.Fatal(srv.ListenAndServe())
}

func connectToDB(ctx context.Context, logger *logging.Logger) balance.Repository {
	cfg := postgresql.DBConfig{
		Username:    os.Getenv("POSTGRES_USERNAME"),
		Password:    os.Getenv("POSTGRES_PASSWORD"),
		Host:        os.Getenv("POSTGRES_HOST"),
		Port:        os.Getenv("POSTGRES_PORT"),
		Database:    os.Getenv("POSTGRES_DATABASE"),
		MaxAttempts: 10,
	}
	postgreSQLClient, err := postgresql.NewClient(ctx, cfg, logger)
	if err != nil {
		log.Fatal(err)
	}
	return db.NewRepository(postgreSQLClient, logger)
}
