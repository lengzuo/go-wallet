package dao

import (
	"context"
	"database/sql"

	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lengzuo/fundflow/configs"
	"github.com/lengzuo/fundflow/pkg/log"
	mysqlx "github.com/lengzuo/fundflow/pkg/sqlx"
	_ "github.com/lib/pq"
)

type DAO struct {
	db *sqlx.DB
}

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

// New creates a new DAO instance with a database connection
func New(ctx context.Context, cfg *configs.DatabaseConfig) (*DAO, error) {
	db, err := mysqlx.New(ctx, cfg.DSN)
	if err != nil {
		log.Error(ctx, "failed to connect to database: %v", err)
		return nil, err
	}
	return &DAO{
		db: db,
	}, nil
}

func lockExecution(ctx context.Context, db *sqlx.DB, fn func(exec sqlx.ExtContext) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error(ctx, "failed to begin transaction: %v", err)
		return fmt.Errorf("begin transaction: %s", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Error(ctx, "failed in txn rollback with err: %s", err)
		}
	}()
	err = fn(tx)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		log.Error(ctx, "failed to commit transaction: %v", err)
		return fmt.Errorf("commit transaction: %s", err)
	}
	return nil
}
