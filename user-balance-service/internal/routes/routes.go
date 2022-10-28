package routes

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/onmono/internal/handler"
	"github.com/onmono/internal/usecases"
	"github.com/onmono/pkg/logging"
	"net/http"
)

func Routes(uc *usecases.UseCase, logger *logging.Logger) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Heartbeat("/api/v1/ping"))

	balanceHandler := handler.NewBalanceHandler(context.TODO(), uc, logger)

	mux.Get("/api/v1/account/balance", balanceHandler.GetBalance)
	mux.Put("/api/v1/account/balance", balanceHandler.DepositOrDebitBalance)
	mux.Post("/api/v1/accounting/reserve", balanceHandler.Reserve)
	mux.Post("/api/v1/accounting/revenue", balanceHandler.Revenue)
	mux.Put("/api/v1/account/money/transfer", balanceHandler.TransferBalance)

	return mux
}
