package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lengzuo/fundflow/internal/apierr"
	"github.com/lengzuo/fundflow/pkg/log"
	"github.com/lengzuo/fundflow/utils"
)

//go:generate mockery --name WalletsRepository --output ./mocks --outpkg mocks --case=underscore
type WalletsRepository interface {
	Deposit(ctx context.Context, username, currency string, amount int) error
	Withdraw(ctx context.Context, username, currency string, amount int) error
	Balance(ctx context.Context, username string, currencies []string) ([]WalletsModel, error)
	Transfer(ctx context.Context, sender, receiver, currency string, amount int) error
	Get(ctx context.Context, username, currency string) (*WalletsModel, error)
}

type WalletsModel struct {
	ID        int       `db:"id"`
	Username  string    `db:"username"`
	Amount    int       `db:"amount"`
	Currency  string    `db:"currency"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type wallets struct {
	db *sqlx.DB
}

func NewWallets(dao *DAO) *wallets {
	return &wallets{
		db: dao.db,
	}
}

func (p *wallets) Deposit(ctx context.Context, username, currency string, amount int) error {
	return lockExecution(ctx, p.db, func(exec sqlx.ExtContext) error {
		transaction := &TransactionsModel{
			UID:         utils.UUID(),
			Type:        TypeDeposit,
			InitiatedBy: username,
			Status:      StatusCompleted,
			Amount:      amount,
			Currency:    currency,
		}
		err := insertTransaction(ctx, exec, transaction)
		if err != nil {
			log.Error(ctx, "failed in insert into transactions with err: %s", err)
			return err
		}
		return updateBalanceAndInsertLedger(ctx, exec, transaction.UID, username, currency, amount, DirectionCredit)
	})
}

func (p *wallets) Withdraw(ctx context.Context, username, currency string, amount int) error {
	return lockExecution(ctx, p.db, func(exec sqlx.ExtContext) error {
		transaction := &TransactionsModel{
			UID:         utils.UUID(),
			Type:        TypeWithdraw,
			InitiatedBy: username,
			Status:      StatusCompleted,
			Amount:      amount,
			Currency:    currency,
		}
		err := insertTransaction(ctx, exec, transaction)
		if err != nil {
			log.Error(ctx, "failed in insert into transactions with err: %s", err)
			return err
		}
		return updateBalanceAndInsertLedger(ctx, exec, transaction.UID, username, currency, amount, DirectionDebit)
	})
}

func (p *wallets) Balance(ctx context.Context, username string, currencies []string) ([]WalletsModel, error) {
	query, args, err := psql.Select("amount", "currency").
		From("wallets").
		Where(squirrel.Eq{
			"username": username,
			"currency": currencies,
		}).
		ToSql()
	if err != nil {
		log.Error(ctx, "failed to build get user balance query with err:%s", err)
		return nil, fmt.Errorf("build get user balance query: %w", err)
	}
	// Assuming we only support JPY and SGD wallet in coding test
	wallets := make([]WalletsModel, 2)
	err = p.db.SelectContext(ctx, &wallets, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierr.NotFound
		}
		log.Error(ctx, "failed in getting user wallet balance with err: %s", err)
		return nil, err
	}

	return wallets, nil
}

func (p *wallets) Transfer(ctx context.Context, sender, receiver, currency string, amount int) error {
	return lockExecution(ctx, p.db, func(exec sqlx.ExtContext) error {
		strs := []string{sender, receiver}
		// Sort the keys to ensure select...for update always in the same sequence for both user, eg, user A transfer to user B and user B transfer to user A at the same time.
		// the locker is able to locked the data properly without causing race condition.
		slices.Sort(strs)

		_, err := get(ctx, exec, strs[0], currency, true)
		if err != nil {
			return err
		}
		_, err = get(ctx, exec, strs[1], currency, true)
		if err != nil {
			return err
		}

		transaction := &TransactionsModel{
			UID:         utils.UUID(),
			Type:        TypeTransfer,
			InitiatedBy: sender,
			Status:      StatusCompleted,
			Amount:      amount,
			Currency:    currency,
		}
		err = insertTransaction(ctx, exec, transaction)
		if err != nil {
			log.Error(ctx, "failed in insert into transactions with err: %s", err)
			return err
		}

		err = updateBalanceAndInsertLedger(ctx, exec, transaction.UID, sender, currency, amount, DirectionDebit)
		if err != nil {
			log.Error(ctx, "failed in update sender and add ledger with err: %s", err)
			return err
		}

		err = updateBalanceAndInsertLedger(ctx, exec, transaction.UID, receiver, currency, amount, DirectionCredit)
		if err != nil {
			log.Error(ctx, "failed in update receiver and add ledger with err: %s", err)
			return err
		}
		return nil
	})
}

func (p *wallets) Get(ctx context.Context, username, currency string) (*WalletsModel, error) {
	return get(ctx, p.db, username, currency, false)
}

func updateBalance(ctx context.Context, exec sqlx.ExtContext, username, currency string, amount int) error {
	if amount == 0 {
		return fmt.Errorf("amount cannot be zero")
	}
	updateBuilder := psql.Update("wallets").
		Set("updated_at", squirrel.Expr("NOW()")).
		Set("amount", squirrel.Expr("amount + ?", amount)).
		Where(squirrel.Eq{
			"username": username,
			"currency": currency,
		})

	if amount < 0 {
		updateBuilder = updateBuilder.Where(squirrel.Expr("amount >= ?", -amount))
	}

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		log.Error(ctx, "failed to build wallet query with err: %s", err)
		return fmt.Errorf("failed to build wallet query: %s", err)
	}
	r, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		log.Error(ctx, "failed to execute wallets update with err: %v", err)
		return fmt.Errorf("failed to execute wallets update: %w", err)
	}
	rowAffected, err := r.RowsAffected()
	if err != nil {
		log.Error(ctx, "unabled to get row affected due to: %s", err)
		return err
	}
	if rowAffected == 0 {
		if amount <= 0 {
			return apierr.InsufficientFund
		}
		return errors.New("unexpected: no rows affected by update")
	}
	log.Debug(ctx, "Wallets updated: %s", username)
	return nil
}

func updateBalanceAndInsertLedger(ctx context.Context, exec sqlx.ExtContext, txUID, username, currency string, amount int, direction Direction) error {
	// Ensure the amount always positive in ledgers table
	ledger := &LedgersModel{
		TxUID:     txUID,
		Username:  username,
		Amount:    amount,
		Currency:  currency,
		Direction: direction,
	}
	// Deduct amount from wallet id if 'debit'
	if direction == DirectionDebit {
		amount = -amount
	}
	err := updateBalance(ctx, exec, username, currency, amount)
	if err != nil {
		return err
	}
	err = insertLedgers(ctx, exec, ledger)
	if err != nil {
		log.Error(ctx, "failed in insert into ledgers with err: %s", err)
		return err
	}
	return nil
}

func get(ctx context.Context, exec sqlx.ExtContext, username, currency string, forUpdate bool) (*WalletsModel, error) {
	queryBuilder := psql.Select("id", "username", "amount").
		From("wallets").
		Where(
			squirrel.And{
				squirrel.Eq{"username": username},
				squirrel.Eq{"currency": currency},
			},
		).
		Limit(1)

	if forUpdate {
		queryBuilder = queryBuilder.Suffix("FOR UPDATE")
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		log.Error(ctx, "failed in build select with err: %s", err)
		return nil, err
	}
	wallet := new(WalletsModel)
	err = exec.QueryRowxContext(ctx, query, args...).StructScan(wallet)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apierr.NotFound
		}
		log.Error(ctx, "failed in get wallet1 %s with err: %s", username, err)
		return nil, err
	}
	return wallet, nil
}
