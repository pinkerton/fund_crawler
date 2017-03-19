package backfill

import (
	"github.com/jmoiron/sqlx"
)

// Type alias for the closure holding the databse reference.
type DB func() *sqlx.DB

// Returns a closure of a DB type.
func Database() DB {
	db := sqlx.MustConnect("sqlite3", "../db/fund_crawler.sqlite")
	return func() *sqlx.DB {
		return db
	}
}
