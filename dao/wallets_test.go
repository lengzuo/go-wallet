package dao

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/lengzuo/fundflow/internal/apierr"
	"github.com/stretchr/testify/assert"
)

func Test_NewWallets(t *testing.T) {
	mockDB, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	t.Run("correct new wallets", func(t *testing.T) {
		daoInstance := &DAO{sqlx.NewDb(mockDB, "sqlmockwallet")}
		walletDao := NewWallets(daoInstance)
		assert.Equal(t, daoInstance.db.DriverName(), walletDao.db.DriverName())
		assert.Implements(t, (*WalletsRepository)(nil), walletDao)
	})
}

func Test_wallets_Deposit(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	p := &wallets{
		db: sqlx.NewDb(mockDB, "sqlmock"),
	}
	t.Run("ok, deposit wallet no error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "deposit", "name", "SGD", 100, "completed", "ref").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3").
			WithArgs(100, "SGD", "name").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO ledgers (tx_uid,username,currency,amount,direction) VALUES ($1,$2,$3,$4,$5)").
			WithArgs(sqlmock.AnyArg(), "name", "SGD", 100, DirectionCredit).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit().WillReturnError(nil)

		err := p.Deposit(t.Context(), "name", "ref", "SGD", 100)
		assert.NoError(t, err, "deposit err")
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("insert deposit ledgers error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "deposit", "name", "SGD", 100, "completed", "ref").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3").
			WithArgs(100, "SGD", "name").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO ledgers (tx_uid,username,currency,amount,direction) VALUES ($1,$2,$3,$4,$5)").
			WithArgs(sqlmock.AnyArg(), "name", "SGD", 100, DirectionCredit).
			WillReturnError(errors.New("err"))

		mock.ExpectRollback().WillReturnError(nil)

		err := p.Deposit(t.Context(), "name", "ref", "SGD", 100)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("update deposit wallets error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "deposit", "name", "SGD", 100, "completed", "ref").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3").
			WithArgs(100, "SGD", "name").
			WillReturnError(errors.New("err"))

		mock.ExpectRollback().WillReturnError(nil)

		err := p.Deposit(t.Context(), "name", "ref", "SGD", 100)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("insert deposit transactions error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "deposit", "name", "SGD", 100, "completed", "ref").
			WillReturnError(errors.New("err"))

		mock.ExpectRollback().WillReturnError(nil)

		err := p.Deposit(t.Context(), "name", "ref", "SGD", 100)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("begin deposit tx error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(errors.New("err"))
		err := p.Deposit(t.Context(), "name", "ref", "SGD", 100)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})
}

func Test_wallets_Withdraw(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	p := &wallets{
		db: sqlx.NewDb(mockDB, "sqlmock"),
	}

	t.Run("ok, withdraw wallet no error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "withdraw", "name2", "SGD", 100, "completed", "ref").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3 AND amount >= $4").
			WithArgs(-100, "SGD", "name2", 100).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO ledgers (tx_uid,username,currency,amount,direction) VALUES ($1,$2,$3,$4,$5)").
			WithArgs(sqlmock.AnyArg(), "name2", "SGD", 100, DirectionDebit).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit().WillReturnError(nil)

		err := p.Withdraw(t.Context(), "name2", "ref", "SGD", 100)
		assert.NoError(t, err, "withdraw err")
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("insert withdraw ledgers error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "withdraw", "name2", "SGD", 100, "completed", "ref").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3 AND amount >= $4").
			WithArgs(-100, "SGD", "name2", 100).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO ledgers (tx_uid,username,currency,amount,direction) VALUES ($1,$2,$3,$4,$5)").
			WithArgs(sqlmock.AnyArg(), "name2", "SGD", 100, DirectionDebit).
			WillReturnError(errors.New("err"))

		mock.ExpectRollback().WillReturnError(nil)

		err := p.Withdraw(t.Context(), "name2", "ref", "SGD", 100)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("update withdraw wallet error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "withdraw", "name2", "SGD", 100, "completed", "ref").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3 AND amount >= $4").
			WithArgs(-100, "SGD", "name2", 100).
			WillReturnError(errors.New("err"))

		mock.ExpectRollback().WillReturnError(nil)

		err := p.Withdraw(t.Context(), "name2", "ref", "SGD", 100)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("insert withdraw transactions error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "withdraw", "name2", "SGD", 101, "completed", "ref").
			WillReturnError(errors.New("err"))

		mock.ExpectRollback().WillReturnError(nil)

		err := p.Withdraw(t.Context(), "name2", "ref", "SGD", 101)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("begin withdraw tx error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(errors.New("err"))
		err := p.Withdraw(t.Context(), "name2", "ref", "SGD", 100)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})
}

func Test_wallets_Balance(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	p := &wallets{
		db: sqlx.NewDb(mockDB, "sqlmock"),
	}

	t.Run("ok get wallet balance", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"amount", "currency"})
		rows.AddRow(20, "SGD")
		rows.AddRow(10, "JPY")
		mock.ExpectQuery("SELECT amount, currency FROM wallets WHERE currency IN ($1,$2) AND username = $3").
			WithArgs("SGD", "JPY", "name").
			WillReturnRows(rows).
			WillReturnError(nil)
		wallets, err := p.Balance(t.Context(), "name", []string{"SGD", "JPY"})
		assert.NoError(t, err)
		assert.Equal(t, wallets[0].Currency, "SGD")
		assert.Equal(t, wallets[0].Amount, 20)
		assert.Equal(t, wallets[1].Currency, "JPY")
		assert.Equal(t, wallets[1].Amount, 10)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("err wallet balance no row return", func(t *testing.T) {
		mock.ExpectQuery("SELECT amount, currency FROM wallets WHERE currency IN ($1) AND username = $2").
			WithArgs("JPY", "name22").
			WillReturnError(sql.ErrNoRows)
		wallets, err := p.Balance(t.Context(), "name22", []string{"JPY"})
		assert.Error(t, err)
		assert.ErrorIs(t, err, apierr.NotFound)
		assert.Nil(t, wallets)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("err wallet balance other", func(t *testing.T) {
		expectedErr := errors.New("err")
		mock.ExpectQuery("SELECT amount, currency FROM wallets WHERE currency IN ($1,$2) AND username = $3").
			WithArgs("JPY", "SGD", "name").
			WillReturnError(expectedErr)
		wallets, err := p.Balance(t.Context(), "name", []string{"JPY", "SGD"})
		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, wallets)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})
}

func Test_wallets_Get(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	p := &wallets{
		db: sqlx.NewDb(mockDB, "sqlmock"),
	}
	t.Run("ok wallet get", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(1, "name", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1").
			WithArgs("name", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)
		wallet, err := p.Get(t.Context(), "name", "SGD")
		assert.NoError(t, err)
		assert.Equal(t, &WalletsModel{
			ID: 1, Username: "name", Amount: 10,
		}, wallet)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("query execute wallet get error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(1, "name", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1").
			WithArgs("name", "SGD").
			WillReturnError(errors.New("err"))
		wallet, err := p.Get(t.Context(), "name", "SGD")
		assert.Error(t, err)
		assert.Nil(t, wallet)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("query execute wallet get no row", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(1, "name", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1").
			WithArgs("name", "SGD").
			WillReturnError(sql.ErrNoRows)
		wallet, err := p.Get(t.Context(), "name", "SGD")
		assert.Error(t, err)
		assert.ErrorIs(t, err, apierr.NotFound)
		assert.Nil(t, wallet)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})
}

func Test_wallets_Transfer(t *testing.T) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	p := &wallets{
		db: sqlx.NewDb(mockDB, "sqlmock"),
	}
	t.Run("ok wallets transfer", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		rows := sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(1, "name1", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name1", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		rows = sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(2, "name2", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name2", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "transfer", "name2", "SGD", 100, "completed", "ref").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3 AND amount >= $4").
			WithArgs(-100, "SGD", "name2", 100).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO ledgers (tx_uid,username,currency,amount,direction) VALUES ($1,$2,$3,$4,$5)").
			WithArgs(sqlmock.AnyArg(), "name2", "SGD", 100, DirectionDebit).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3").
			WithArgs(100, "SGD", "name1").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO ledgers (tx_uid,username,currency,amount,direction) VALUES ($1,$2,$3,$4,$5)").
			WithArgs(sqlmock.AnyArg(), "name1", "SGD", 100, DirectionCredit).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit().WillReturnError(nil)

		err = p.Transfer(t.Context(), "name2", "name1", "ref", "SGD", 100)
		assert.NoError(t, err, "transfer err")
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("reverse name1 and name2 for update still need to be same sequence", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		rows := sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(1, "name1", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name1", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		rows = sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(2, "name2", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name2", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "transfer", "name1", "SGD", 100, "completed", "ref").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3 AND amount >= $4").
			WithArgs(-100, "SGD", "name1", 100).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO ledgers (tx_uid,username,currency,amount,direction) VALUES ($1,$2,$3,$4,$5)").
			WithArgs(sqlmock.AnyArg(), "name1", "SGD", 100, DirectionDebit).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3").
			WithArgs(100, "SGD", "name2").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO ledgers (tx_uid,username,currency,amount,direction) VALUES ($1,$2,$3,$4,$5)").
			WithArgs(sqlmock.AnyArg(), "name2", "SGD", 100, DirectionCredit).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit().WillReturnError(nil)

		err = p.Transfer(t.Context(), "name1", "name2", "ref", "SGD", 100)
		assert.NoError(t, err, "transfer err")
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("insert receiver ledgers failed wallets transfer", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		rows := sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(1, "name1", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name1", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		rows = sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(2, "name2", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name2", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "transfer", "name1", "SGD", 10, "completed", "ref").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3 AND amount >= $4").
			WithArgs(-10, "SGD", "name1", 10).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO ledgers (tx_uid,username,currency,amount,direction) VALUES ($1,$2,$3,$4,$5)").
			WithArgs(sqlmock.AnyArg(), "name1", "SGD", 10, DirectionDebit).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3").
			WithArgs(10, "SGD", "name2").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO ledgers (tx_uid,username,currency,amount,direction) VALUES ($1,$2,$3,$4,$5)").
			WithArgs(sqlmock.AnyArg(), "name2", "SGD", 10, DirectionCredit).
			WillReturnError(errors.New("err"))

		mock.ExpectRollback().WillReturnError(nil)

		err = p.Transfer(t.Context(), "name1", "name2", "ref", "SGD", 10)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("update receiver wallets failed wallets transfer", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		rows := sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(1, "name1", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name1", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		rows = sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(2, "name2", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name2", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "transfer", "name1", "SGD", 10, "completed", "ref").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3 AND amount >= $4").
			WithArgs(-10, "SGD", "name1", 10).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO ledgers (tx_uid,username,currency,amount,direction) VALUES ($1,$2,$3,$4,$5)").
			WithArgs(sqlmock.AnyArg(), "name1", "SGD", 10, DirectionDebit).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3").
			WithArgs(10, "SGD", "name2").
			WillReturnError(errors.New("err"))

		mock.ExpectRollback().WillReturnError(nil)

		err = p.Transfer(t.Context(), "name1", "name2", "ref", "SGD", 10)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("insert sender ledgers failed wallets transfer", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		rows := sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(1, "name1", 11)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name1", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		rows = sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(2, "name2", 11)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name2", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "transfer", "name1", "SGD", 11, "completed", "ref").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3 AND amount >= $4").
			WithArgs(-11, "SGD", "name1", 11).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("INSERT INTO ledgers (tx_uid,username,currency,amount,direction) VALUES ($1,$2,$3,$4,$5)").
			WithArgs(sqlmock.AnyArg(), "name1", "SGD", 11, DirectionDebit).
			WillReturnError(errors.New("err"))

		mock.ExpectRollback().WillReturnError(nil)

		err = p.Transfer(t.Context(), "name1", "name2", "ref", "SGD", 11)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("update sender wallets failed wallets transfer", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		rows := sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(1, "name1", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name1", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		rows = sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(2, "name2", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name2", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "transfer", "name1", "SGD", 10, "completed", "ref").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec("UPDATE wallets SET updated_at = NOW(), amount = amount + $1 WHERE currency = $2 AND username = $3 AND amount >= $4").
			WithArgs(-10, "SGD", "name1", 10).
			WillReturnError(errors.New("err"))

		mock.ExpectRollback().WillReturnError(nil)

		err = p.Transfer(t.Context(), "name1", "name2", "ref", "SGD", 10)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("insert transactions failed wallets transfer", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)

		rows := sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(1, "name1", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name1", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		rows = sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(2, "name2", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name2", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status,reference) VALUES ($1,$2,$3,$4,$5,$6,$7)").
			WithArgs(sqlmock.AnyArg(), "transfer", "name2", "SGD", 10, "completed", "ref").
			WillReturnError(errors.New("err"))

		mock.ExpectRollback().WillReturnError(nil)

		err = p.Transfer(t.Context(), "name2", "name1", "ref", "SGD", 10)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("select second wallet error wallets transfer", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)
		rows := sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(1, "name1", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name1", "SGD").
			WillReturnRows(rows).
			WillReturnError(nil)

		rows = sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(2, "name2", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name2", "SGD").
			WillReturnError(errors.New("err"))
		mock.ExpectRollback().WillReturnError(nil)
		err = p.Transfer(t.Context(), "name2", "name1", "ref", "SGD", 10)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})

	t.Run("select first wallet error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(nil)
		rows := sqlmock.NewRows([]string{"id", "username", "amount"})
		rows.AddRow(1, "name1", 10)
		mock.ExpectQuery("SELECT id, username, amount FROM wallets WHERE (username = $1 AND currency = $2) LIMIT 1 FOR UPDATE").
			WithArgs("name1", "SGD").
			WillReturnError(errors.New("err"))
		mock.ExpectRollback().WillReturnError(nil)
		err = p.Transfer(t.Context(), "name2", "name1", "ref", "SGD", 10)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
	})
}
