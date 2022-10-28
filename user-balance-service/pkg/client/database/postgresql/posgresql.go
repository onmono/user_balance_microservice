package postgresql

import (
	"context"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/onmono/pkg/logging"
	repeatable "github.com/onmono/pkg/utils"
	"log"
	"time"
)

type DBConfig struct {
	Username    string
	Password    string
	Host        string
	Port        string
	Database    string
	MaxAttempts int
}

type Client interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
}

func NewClient(ctx context.Context, dbConfig DBConfig, logger *logging.Logger) (pool *pgxpool.Pool, err error) {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", dbConfig.Username, dbConfig.Password,
		dbConfig.Host, dbConfig.Port, dbConfig.Database)

	err = repeatable.DoWithTries(func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		pool, err = pgxpool.Connect(ctx, dsn)
		if err != nil {
			logger.Error(err)
			return err
		}

		logger.Infof("Conn String : %s, Total Conns: %v", pool.Config().ConnString(), pool.Stat().TotalConns())

		return nil
	}, dbConfig.MaxAttempts, 5*time.Second)

	if err != nil {
		logger.Errorf("error do with tries postgresql, %v", err)
		log.Fatal("error do with tries postgresql")
	}

	return pool, nil
}
