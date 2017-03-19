package models

import (
	"time"
)

// Fund model
type Fund struct {
	AutoIncr
	Created   time.Time
	Symbol    string
	Name      string
	Type      string
	Available bool
}

// Get a Fund by ID.
func (self *Fund) GetByID(db backfill.DB, id int) (err error) {
	sql := `SELECT * FROM funds WHERE id = ? LIMIT 1;`
	err = db().Get(self, sql, id)
	return
}
