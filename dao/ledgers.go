package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lengzuo/fundflow/pkg/log"
)

//go:generate mockery --name LedgersRepository --output ./mocks --outpkg mocks --case=underscore
type LedgersRepository interface {
	Insert(ctx context.Context, user *LedgersModel) error
	List(ctx context.Context, limit int, startingAfter, currency, username string) ([]TxHistoryModel, bool, error)
}

type Direction string

const (
	// 'c' indicate for credit
	DirectionCredit Direction = "c"
	// 'd' indicate for debit
	DirectionDebit Direction = "d"
)

type ledgers struct {
	db *sqlx.DB
}

func NewLedgers(dao *DAO) *ledgers {
	return &ledgers{
		db: dao.db,
	}
}

type LedgersModel struct {
	ID        int       `db:"id"`
	TxUID     string    `db:"tx_uid"`
	Username  string    `db:"username"`
	Amount    int       `db:"amount"`
	Currency  string    `db:"currency"`
	Direction Direction `db:"direction"`
	CreatedAt time.Time `db:"created_at"`
}

type TxHistoryModel struct {
	UID       string    `db:"uid"`
	Type      TxType    `db:"type"`
	Status    TxStatus  `db:"status"`
	Direction Direction `db:"direction"`
	Amount    int       `db:"amount"`
	Currency  string    `db:"currency"`
	CreatedAt time.Time `db:"created_at"`
}

func (p *ledgers) Insert(ctx context.Context, ledger *LedgersModel) error {
	return insertLedgers(ctx, p.db, ledger)
}

func (p *ledgers) List(ctx context.Context, limit int, startingAfter, currency, username string) ([]TxHistoryModel, bool, error) {
	sq := psql.Select(
		"t.uid",
		"t.type",
		"t.status",
		"le.direction",
		"le.amount",
		"le.currency",
		"le.created_at",
	).
		From("ledgers le").
		Join("transactions t ON t.uid = le.tx_uid").
		Where(squirrel.Eq{
			"le.username": username,
			"le.currency": currency,
		}).
		OrderBy("le.created_at DESC").
		Limit(uint64(limit + 1))

	if startingAfter != "" {
		sq = sq.Where("t.uid < ?", startingAfter)
	}
	txQuery, txArgs, err := sq.ToSql()
	if err != nil {
		log.Error(ctx, "failed to build tx list query: %v", err)
		return nil, false, fmt.Errorf("build tx list query: %w", err)
	}
	log.Debug(ctx, "[query]: %s [args]: %s", txQuery, txArgs)
	txHistories := []TxHistoryModel{}
	err = p.db.SelectContext(ctx, &txHistories, txQuery, txArgs...)
	if err != nil {
		log.Error(ctx, "failed to list tx hisotries: %v", err)
		return nil, false, fmt.Errorf("list tx histories: %w", err)
	}
	hasMore := len(txHistories) > limit
	if hasMore {
		txHistories = txHistories[:limit]
	}
	return txHistories, hasMore, nil
}

func insertLedgers(ctx context.Context, exec sqlx.ExtContext, ledger *LedgersModel) error {
	query, args, err := psql.Insert("ledgers").
		Columns("tx_uid", "username", "currency", "amount", "direction").
		Values(ledger.TxUID, ledger.Username, ledger.Currency, ledger.Amount, ledger.Direction).
		ToSql()
	if err != nil {
		log.Error(ctx, "failed to build ledger insert query: %v", err)
		return fmt.Errorf("build ledger insert query: %w", err)
	}

	_, err = exec.ExecContext(ctx, query, args...)
	if err != nil {
		log.Error(ctx, "failed to insert ledger: %v", err)
		return fmt.Errorf("insert ledger: %w", err)
	}
	log.Debug(ctx, "Ledgers created : %s", ledger.TxUID)
	return nil
}
