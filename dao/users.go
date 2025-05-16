package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lengzuo/fundflow/pkg/log"
	"github.com/lib/pq"
)

type UserRepository interface {
	Insert(ctx context.Context, user *UsersModel) error
}

type UsersModel struct {
	ID        int       `db:"id"`
	Username  string    `db:"username"`
	Password  string    `db:"password"`
	Active    bool      `db:"active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type users struct {
	db *sqlx.DB
}

func NewUsers(dao *DAO) *users {
	return &users{
		db: dao.db,
	}
}

func (p *users) Insert(ctx context.Context, user *UsersModel) error {
	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Error(ctx, "failed to begin transaction: %v", err)
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.Error(ctx, "failed in txn rollback")
		}
	}()

	userQuery, userArgs, err := psql.Insert("users").
		Columns("username", "password", "active").
		Values(user.Username, user.Password, true).
		ToSql()

	if err != nil {
		log.Error(ctx, "failed to build user insert query: %v", err)
		return fmt.Errorf("build user insert query: %w", err)
	}

	_, err = tx.ExecContext(ctx, userQuery, userArgs...)
	if err != nil {
		log.Error(ctx, "failed to insert user: %v", err)
		var pqErr *pq.Error
		// Check if the error is a pq.Error and if its code is '23505' (unique_violation)
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			log.Error(ctx, "user already exists")
			return ErrAlreadyExists
		}
		return fmt.Errorf("insert user: %w", err)
	}

	log.Debug(ctx, "User created successfully : %s", user.Username)

	walletQuery, walletArgs, err := psql.Insert("wallets").
		Columns("username", "currency", "amount").
		Values(user.Username, "SGD", 0).
		ToSql()
	if err != nil {
		log.Error(ctx, "failed to build wallet insert query: %v", err)
		return fmt.Errorf("build wallet insert query: %w", err)
	}

	_, err = tx.ExecContext(ctx, walletQuery, walletArgs...)
	if err != nil {
		log.Error(ctx, "failed to insert default wallet: %v", err)
		return fmt.Errorf("insert default wallet: %w", err)
	}
	// If all operations were successful, commit the transaction
	if err = tx.Commit(); err != nil {
		log.Error(ctx, "failed to commit transaction: %v", err)
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
