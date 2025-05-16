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
		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status) VALUES ($1,$2,$3,$4,$5,$6)").
			WithArgs("uid", "type", "initiator", "SGD", 100, "completed").
			WillReturnResult(sqlmock.NewResult(1, 1)).
			WillReturnError(nil)
		err := p.Insert(t.Context(), &TransactionsModel{
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
		mock.ExpectExec("INSERT INTO transactions (uid,type,initiated_by,currency,amount,status) VALUES ($1,$2,$3,$4,$5,$6)").
			WithArgs("uid", "type", "initiator", "SGD", 100, "completed").
			WillReturnError(errors.New("err"))
		err := p.Insert(t.Context(), &TransactionsModel{
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

// func Test_transactions_List(t *testing.T) {
// 	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
// 	if err != nil {
// 		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
// 	}
// 	defer mockDB.Close()
// 	type args struct {
// 		fn            func(mock sqlmock.Sqlmock)
// 		limit         int
// 		startingAfter string
// 		currency      string
// 		username      string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    []TransactionsModel
// 		want1   bool
// 		wantErr bool
// 	}{
// 		{
// 			name: "ok, startingAfter is empty",
// 			args: args{
// 				fn: func(mock sqlmock.Sqlmock) {
// 					rows := sqlmock.NewRows([]string{"uid", "type", "from_account", "status", "amount", "currency", "created_at"})
// 					rows.AddRow("tx_uid_1", "DEPOSIT", "user123", "COMPLETED", 100, "SGD", time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC))
// 					rows.AddRow("tx_uid_2", "WITHDRAW", "user1", "COMPLETED", 99, "SGD", time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC))

// 					mock.ExpectQuery("SELECT uid, type, from_account, status, amount, currency, created_at FROM transactions WHERE ((from_account = $1 OR to_account = $2) AND currency = $3) ORDER BY created_at DESC LIMIT 2").
// 						WithArgs("user123", "user123", "SGD").
// 						WillReturnRows(rows).
// 						RowsWillBeClosed().
// 						WillReturnError(nil)
// 				},
// 				limit:    1,
// 				currency: "SGD",
// 				username: "user123",
// 			},
// 			want: []TransactionsModel{
// 				{
// 					UID:         "tx_uid_1",
// 					Type:        TypeDeposit,
// 					FromAccount: "user123",
// 					Status:      StatusCompleted,
// 					Amount:      100,
// 					Currency:    "SGD",
// 					CreatedAt:   time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC),
// 				},
// 			},
// 			want1:   true,
// 			wantErr: false,
// 		},
// 		{
// 			name: "ok, startingAfter is not empty",
// 			args: args{
// 				fn: func(mock sqlmock.Sqlmock) {
// 					rows := sqlmock.NewRows([]string{"uid", "type", "from_account", "status", "amount", "currency", "created_at"})
// 					rows.AddRow("tx_uid_1", "DEPOSIT", "user123", "COMPLETED", 100, "SGD", time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC))
// 					rows.AddRow("tx_uid_2", "WITHDRAW", "user1", "COMPLETED", 99, "SGD", time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC))

// 					mock.ExpectQuery("SELECT uid, type, from_account, status, amount, currency, created_at FROM transactions WHERE ((from_account = $1 OR to_account = $2) AND currency = $3) AND uid > $4 ORDER BY created_at DESC LIMIT 3").
// 						WithArgs("user123", "user123", "SGD", "tx_id").
// 						WillReturnRows(rows).
// 						RowsWillBeClosed().
// 						WillReturnError(nil)
// 				},
// 				startingAfter: "tx_id",
// 				limit:         2,
// 				currency:      "SGD",
// 				username:      "user123",
// 			},
// 			want: []TransactionsModel{
// 				{
// 					UID:         "tx_uid_1",
// 					Type:        TypeDeposit,
// 					FromAccount: "user123",
// 					Status:      StatusCompleted,
// 					Amount:      100,
// 					Currency:    "SGD",
// 					CreatedAt:   time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC),
// 				},
// 				{
// 					UID:         "tx_uid_2",
// 					Type:        TypeWithdraw,
// 					FromAccount: "user1",
// 					Status:      StatusCompleted,
// 					Amount:      99,
// 					Currency:    "SGD",
// 					CreatedAt:   time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC),
// 				},
// 			},
// 			want1:   false,
// 			wantErr: false,
// 		},
// 		{
// 			name: "error, in select",
// 			args: args{
// 				fn: func(mock sqlmock.Sqlmock) {
// 					mock.ExpectQuery("SELECT uid, type, from_account, status, amount, currency, created_at FROM transactions WHERE ((from_account = $1 OR to_account = $2) AND currency = $3) AND uid > $4 ORDER BY created_at DESC LIMIT 3").
// 						WithArgs("user123", "user123", "SGD", "tx_id").
// 						WillReturnError(errors.New("err"))
// 				},
// 				startingAfter: "tx_id",
// 				limit:         2,
// 				currency:      "SGD",
// 				username:      "user123",
// 			},
// 			want:    nil,
// 			want1:   false,
// 			wantErr: true,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			p := &transactions{
// 				db: sqlx.NewDb(mockDB, "sqlmock"),
// 			}
// 			tt.args.fn(mock)
// 			got, got1, err := p.List(t.Context(), tt.args.limit, tt.args.startingAfter, tt.args.currency, tt.args.username)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("transactions.List() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("transactions.List() got = %v, want %v", got, tt.want)
// 			}
// 			if got1 != tt.want1 {
// 				t.Errorf("transactions.List() got1 = %v, want %v", got1, tt.want1)
// 			}
// 			if err := mock.ExpectationsWereMet(); err != nil {
// 				t.Errorf("there were unfulfilled expectations: %s", err)
// 			}
// 		})
// 	}
// }
