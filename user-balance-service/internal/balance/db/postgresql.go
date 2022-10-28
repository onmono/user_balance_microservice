package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/onmono/internal/balance"
	"github.com/onmono/internal/balance/models"
	"github.com/onmono/pkg/client/database/postgresql"
	"github.com/onmono/pkg/logging"
	"time"
)

type repository struct {
	client postgresql.Client
	logger *logging.Logger
}

func NewRepository(client postgresql.Client, logger *logging.Logger) balance.Repository {
	return &repository{
		client: client,
		logger: logger,
	}
}

func (r *repository) GetConn(ctx context.Context) (*pgxpool.Conn, error) {
	conn, err := r.client.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func (r *repository) Create(ctx context.Context, model models.UserBalance) (*balance.ConnTx, error) {
	conn, err := r.client.Acquire(ctx)
	if err != nil {
		return &balance.ConnTx{Conn: conn, Tx: nil}, err
	}

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return &balance.ConnTx{Conn: conn, Tx: tx}, err
	}
	q := `
	INSERT INTO user_balance (id,user_id,balance,last_updated_at)
	VALUES ($1,$2,$3,$4)
	RETURNING id
	`
	model.ID = uuid.New()

	if err = tx.QueryRow(ctx, q, model.ID, model.UserID, model.Balance, time.Now().UTC()).Scan(&model.UserID); err != nil {
		var pgErr *pgconn.PgError
		if errors.Is(err, pgErr) {
			r.logger.Error(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))

			return &balance.ConnTx{Conn: conn, Tx: tx}, pgErr
		}
		r.logger.Error(err.Error())
		return &balance.ConnTx{Conn: conn, Tx: tx}, err
	}
	return &balance.ConnTx{Conn: conn, Tx: tx}, nil
}

func (r *repository) FindOne(ctx context.Context, id uuid.UUID) (model models.UserBalance, err error) {
	conn, err := r.client.Acquire(ctx)
	if err != nil {
		return models.UserBalance{}, err
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return model, err
	}

	q := `
		SELECT id, user_id, balance, last_updated_at FROM user_balance WHERE user_id = $1;
	`

	if err = tx.QueryRow(ctx, q, id).Scan(&model.ID, &model.UserID, &model.Balance, &model.LastUpdatedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.Is(err, pgErr) {
			r.logger.Error(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
			return model, pgErr
		}
		r.logger.Error(err.Error())
		return model, err
	}
	return model, nil
}

func (r *repository) Update(ctx context.Context, in models.UserBalance) (*balance.ConnTx, error) {
	conn, err := r.client.Acquire(ctx)
	if err != nil {
		return &balance.ConnTx{Conn: conn, Tx: nil}, err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return &balance.ConnTx{Conn: conn, Tx: tx}, err
	}
	q := `
		UPDATE user_balance
		SET balance = $2, last_updated_at = $3
		WHERE user_id = $1
		RETURNING balance;
	`
	in.LastUpdatedAt = time.Now().UTC()
	err = tx.QueryRow(ctx, q, in.UserID, in.Balance, in.LastUpdatedAt).Scan(&in.Balance)
	if err != nil {
		r.logger.Error(err.Error())
		return &balance.ConnTx{Conn: conn, Tx: tx}, err
	}
	return &balance.ConnTx{Conn: conn, Tx: tx}, err
}

func (r *repository) Reserve(ctx context.Context, in models.Reserve) (*balance.ConnTx, error) {
	conn, err := r.client.Acquire(ctx)
	if err != nil {
		return &balance.ConnTx{Conn: conn, Tx: nil}, err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	q := `
		INSERT INTO reserve_info (id,reserve_id,user_id,service_id,order_id,price,timestamp) VALUES ($1,$2,$3,$4,$5,$6,$7);
	`
	in.LastUpdatedAt = time.Now().UTC()
	_, err = tx.Exec(ctx, q, in.ID, in.ReserveID, in.UserID, in.ServiceID, in.OrderID, in.Price, in.LastUpdatedAt)
	if err != nil {
		r.logger.Error(err.Error())
		return &balance.ConnTx{Conn: conn, Tx: tx}, err
	}
	return &balance.ConnTx{Conn: conn, Tx: tx}, err
}

func (r *repository) DeleteReserve(ctx context.Context, id uuid.UUID) error {
	conn, err := r.client.Acquire(ctx)
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	query := `
		DELETE FROM reserve_info WHERE id = $1;
	`
	_, err = tx.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	tx.Commit(ctx)
	conn.Release()
	return nil
}

func (r *repository) CreateRevenue(ctx context.Context, in models.AccountingRevenue) (*balance.ConnTx, error) {
	conn, err := r.client.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	q := `
	INSERT INTO accounting_revenue (id,user_id,service_id,order_id,sum,timestamp)
	VALUES ($1,$2,$3,$4,$5,$6)
	RETURNING id
	`
	if _, err = tx.Exec(ctx, q, in.ID, in.UserID, in.ServiceID, in.OrderID, in.Sum, in.Timestamp); err != nil {
		var pgErr *pgconn.PgError
		if errors.Is(err, pgErr) {
			r.logger.Error(fmt.Sprintf("SQL Error: %s, Detail: %s, Where: %s, Code: %s, SQLState: %s",
				pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))

			return &balance.ConnTx{Conn: conn, Tx: tx}, pgErr
		}
		r.logger.Error(err.Error())
		return &balance.ConnTx{Conn: conn, Tx: tx}, err
	}
	return &balance.ConnTx{Conn: conn, Tx: tx}, nil
}

func (r *repository) DeleteUserBalance(ctx context.Context, id uuid.UUID) error {
	conn, err := r.client.Acquire(ctx)
	if err != nil {
		return err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	query := `
		DELETE FROM user_balance WHERE user_id = $1;
	`
	_, err = tx.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	err = tx.Commit(ctx)
	conn.Release()
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) GetReserve(ctx context.Context, in models.Reserve) ([]models.Reserve, error) {
	conn, err := r.client.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadWrite,
	})
	// сначала select, выбрать rows, попробовать снять с баланса пользователя в цикле по всем заявкам
	selectQuery := `
		SELECT * FROM reserve_info WHERE user_id = $1 AND service_id = $2 AND order_id = $3 AND price = $4
	`
	selectRows, err := tx.Query(ctx, selectQuery, in.UserID, in.ServiceID, in.OrderID, in.Price)
	if err != nil {
		return nil, err
	}

	reserves := make([]models.Reserve, 0, 1)

	for selectRows.Next() {
		model := models.Reserve{}
		//selectRows.Scan(&model.ID, &model.ReserveID, &model.UserID, &model.ServiceID,
		//	&model.OrderID, &model.Price, &model.LastUpdatedAt)

		err = selectRows.Scan(&model.ID, &model.ReserveID, &model.UserID, &model.ServiceID,
			&model.OrderID, &model.Price, &model.LastUpdatedAt)
		if err != nil {
			return nil, err
		}
		reserves = append(reserves, model)
	}
	return reserves, err
}
