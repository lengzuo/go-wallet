package sqlx

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/lengzuo/fundflow/pkg/log"
)

func New(ctx context.Context, dsn string) (*sqlx.DB, error) {
	// Connect to PostgreSQL using the URI format
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Error(ctx, "failed to connect to database: %v", err)
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	err = db.Ping()
	if err != nil {
		log.Error(ctx, "failed in db ping")
	}
	return db, nil
}
