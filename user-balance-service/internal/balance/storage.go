package balance

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/onmono/internal/balance/models"
)

type ConnTx struct {
	Conn *pgxpool.Conn
	Tx   pgx.Tx
}

type Repository interface {
	Create(ctx context.Context, model models.UserBalance) (*ConnTx, error)
	FindOne(ctx context.Context, id uuid.UUID) (model models.UserBalance, err error)
	Update(ctx context.Context, in models.UserBalance) (*ConnTx, error)
	Reserve(ctx context.Context, in models.Reserve) (*ConnTx, error)
	CreateRevenue(ctx context.Context, in models.AccountingRevenue) (*ConnTx, error)
	GetReserve(ctx context.Context, in models.Reserve) ([]models.Reserve, error)
	DeleteReserve(ctx context.Context, id uuid.UUID) error
	DeleteUserBalance(ctx context.Context, id uuid.UUID) error
}
