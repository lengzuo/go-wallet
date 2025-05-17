package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lengzuo/fundflow/pkg/log"
)

//go:generate mockery --name TransactionsRepository --output ./mocks --outpkg mocks --case=underscore
type TransactionsRepository interface {
	Insert(ctx context.Context, user TransactionsModel) error
	ListByReference(ctx context.Context, limit int, startingAfter, reference, username string) ([]TransactionsModel, bool, error)
	GetByUID(ctx context.Context, uid string) (*TransactionsModel, error)
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
	Reference   string    `db:"reference"`
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

func buildSelect() squirrel.SelectBuilder {
	return psql.Select("reference", "initiated_by", "uid", "type", "status", "amount", "currency", "created_at").
		From("transactions")
}

func (p *transactions) Insert(ctx context.Context, tx TransactionsModel) error {
	return insertTransaction(ctx, p.db, tx)
}

func (p *transactions) ListByReference(ctx context.Context, limit int, startingAfter, reference, username string) ([]TransactionsModel, bool, error) {
	sq := buildSelect().
		Where(squirrel.Eq{
			"reference":    reference,
			"initiated_by": username,
		}).
		OrderBy("created_at DESC").
		Limit(uint64(limit + 1))
	if startingAfter != "" {
		sq = sq.Where("uid < ?", startingAfter)
	}
	query, args, err := sq.ToSql()
	if err != nil {
		log.Error(ctx, "failed to build list txn by reference with err: %s", err)
		return nil, false, err
	}
	transactions := []TransactionsModel{}
	err = p.db.SelectContext(ctx, &transactions, query, args...)
	if err != nil {
		log.Error(ctx, "failed to list tx from reference: %v", err)
		return nil, false, fmt.Errorf("list tx from reference: %s", err)
	}
	hasMore := len(transactions) > limit
	if hasMore {
		transactions = transactions[:limit]
	}
	return transactions, hasMore, nil
}

func (p *transactions) GetByUID(ctx context.Context, uid string) (*TransactionsModel, error) {
	query, args, err := buildSelect().
		Where(squirrel.Eq{
			"uid": uid,
		}).
		ToSql()
	if err != nil {
		log.Error(ctx, "failed to build get txn by uid err: %s", err)
		return nil, err
	}
	transaction := new(TransactionsModel)
	err = p.db.GetContext(ctx, transaction, query, args...)
	if err != nil {
		log.Error(ctx, "failed to get tx by uid: %v", err)
		return nil, fmt.Errorf("failed to get tx by uid: %s", err)
	}
	return transaction, nil
}

func insertTransaction(ctx context.Context, exec sqlx.ExtContext, tx TransactionsModel) error {
	txQuery, txArgs, err := psql.Insert("transactions").
		Columns("uid", "type", "initiated_by", "currency", "amount", "status", "reference").
		Values(tx.UID, tx.Type, tx.InitiatedBy, tx.Currency, tx.Amount, tx.Status, tx.Reference).
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
