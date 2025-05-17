package dao

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestNewTransactions(t *testing.T) {
	mockDB, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	t.Run("correct init", func(t *testing.T) {
		daoInstance := &DAO{sqlx.NewDb(mockDB, "sqlmock")}
		transactionDAO := NewTransactions(daoInstance)
		assert.Equal(t, daoInstance.db.DriverName(), transactionDAO.db.DriverName())
		assert.Implements(t, (*TransactionsRepository)(nil), transactionDAO)
	})
}

func Test_transactions_Insert(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	p := &transactions{
		db: sqlx.NewDb(mockDB, "sqlmock"),
	}
	t.Run("ok transaction insert", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs("uid", "type", "initiator", "SGD", 100, "completed", "ref").
			WillReturnResult(sqlmock.NewResult(1, 1)).
			WillReturnError(nil)
		err := p.Insert(t.Context(), TransactionsModel{
			Reference:   "ref",
			UID:         "uid",
			Type:        "type",
			InitiatedBy: "initiator",
			Status:      "completed",
			Amount:      100,
			Currency:    "SGD",
			Metadata:    "",
		})
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("error transaction insert", func(t *testing.T) {
		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs("uid", "type", "initiator", "SGD", 100, "completed", "ref").
			WillReturnError(errors.New("err"))
		err := p.Insert(t.Context(), TransactionsModel{
			Reference:   "ref",
			UID:         "uid",
			Type:        "type",
			InitiatedBy: "initiator",
			Status:      "completed",
			Amount:      100,
			Currency:    "SGD",
			Metadata:    "",
		})
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})
}
