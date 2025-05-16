package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lengzuo/fundflow/pkg/log"
)

//go:generate mockery --name TransactionsRepository --output ./mocks --outpkg mocks --case=underscore
type TransactionsRepository interface {
	Insert(ctx context.Context, user *TransactionsModel) error
	// List(ctx context.Context, limit int, startingAfter, currency, username string) ([]TransactionsModel, bool, error)
}

type TxType string

const (
	TypeDeposit  TxType = "deposit"
	TypeWithdraw TxType = "withdraw"
	TypeTransfer TxType = "transfer"
)

type TxStatus string

const (
	StatusPending   TxStatus = "pending"
	StatusCompleted TxStatus = "completed"
	StatusFailed    TxStatus = "failed"
	StatusCancelled TxStatus = "cancelled"
)

type TransactionsModel struct {
	ID          int       `db:"id"`
	UID         string    `db:"uid"`
	Type        TxType    `db:"type"`
	InitiatedBy string    `db:"initiated_by"`
	Status      TxStatus  `db:"status"`
	Amount      int       `db:"amount"`
	Currency    string    `db:"currency"`
	Metadata    string    `db:"metadata"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type transactions struct {
	db *sqlx.DB
}

func NewTransactions(dao *DAO) *transactions {
	return &transactions{
		db: dao.db,
	}
}

func (p *transactions) Insert(ctx context.Context, tx *TransactionsModel) error {
	return insertTransaction(ctx, p.db, tx)
}

// func (p *transactions) List(ctx context.Context, limit int, startingAfter, currency, username string) ([]TransactionsModel, bool, error) {
// 	sq := psql.Select("uid", "type", "from_account", "status", "amount", "currency", "created_at").
// 		From("transactions").
// 		Where(
// 			squirrel.And{
// 				squirrel.Or{
// 					squirrel.Eq{"from_account": username},
// 					squirrel.Eq{"to_account": username},
// 				},
// 				squirrel.Eq{"currency": currency},
// 			}).
// 		OrderBy("created_at DESC").
// 		Limit(uint64(limit) + 1)

// 	if startingAfter != "" {
// 		sq = sq.Where("uid > ?", startingAfter)
// 	}

// 	txQuery, txArgs, err := sq.ToSql()
// 	if err != nil {
// 		log.Error(ctx, "failed to build tx list query: %v", err)
// 		return nil, false, fmt.Errorf("build tx list query: %w", err)
// 	}

// 	transactions := []TransactionsModel{}
// 	err = p.db.SelectContext(ctx, &transactions, txQuery, txArgs...)
// 	if err != nil {
// 		log.Error(ctx, "failed to list transactions: %v", err)
// 		return nil, false, fmt.Errorf("list transactions: %w", err)
// 	}
// 	hasMore := len(transactions) > limit
// 	if hasMore {
// 		transactions = transactions[:limit]
// 	}
// 	return transactions, hasMore, nil
// }

func insertTransaction(ctx context.Context, exec sqlx.ExtContext, tx *TransactionsModel) error {
	txQuery, txArgs, err := psql.Insert("transactions").
		Columns("uid", "type", "initiated_by", "currency", "amount", "status").
		Values(tx.UID, tx.Type, tx.InitiatedBy, tx.Currency, tx.Amount, tx.Status).
		ToSql()
	if err != nil {
		log.Error(ctx, "failed to build tx insert query: %v", err)
		return fmt.Errorf("build tx insert query: %s", err)
	}
	_, err = exec.ExecContext(ctx, txQuery, txArgs...)
	if err != nil {
		log.Error(ctx, "failed to insert transactions: %v", err)
		return fmt.Errorf("insert transactions: %s", err)
	}
	log.Debug(ctx, "Transactions created : %s", tx.UID)
	return nil
}
