package dao

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestNewLedgers(t *testing.T) {
	mockDB, _, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	t.Run("correct init", func(t *testing.T) {
		daoInstance := &DAO{sqlx.NewDb(mockDB, "sqlmock")}
		ledgerDAO := NewLedgers(daoInstance)
		assert.Equal(t, daoInstance.db.DriverName(), ledgerDAO.db.DriverName())
		assert.Implements(t, (*LedgersRepository)(nil), ledgerDAO)
	})
}
