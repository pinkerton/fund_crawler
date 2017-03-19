package models

// Record model
type Record struct {
	AutoIncr
	Created  time.Time
	Day      time.Time // TODO get right day
	Open     real      // TODO get right type
	Close    real
	RecordID uint64 `db:"record_id"`
	Record   `db:"records"`
}

// Create an R.
func (self *Record) Create(db backfill.DB, day time.Time, open_price real, close_price real, u *User) (err error) {
	insertSQL := `INSERT INTO
                    records (day, open, close) 
                VALUES (?, ?, ?, ?, ?);`
	selectSQL := `SELECT 
                    records.*,
                    users.id "users.id",
                    users.created "users.created",
                    users.phone_number "users.phone_number",
                    users.sleeping "users.sleeping",
                    users.frequency_hours "users.frequency_hours"
                FROM
                    records
                JOIN
                    users ON records.user_id = users.id
                AND
                    records.id = ?
                LIMIT 1;`
	// Transaction to INSERT an record and SELECT its row
	tx := db().MustBegin()
	result := tx.MustExec(insertSQL, start, end, title)
	id, err := result.LastInsertId()
	if err != nil {
		// Something went wrong inserting the row
		tx.Rollback()
		return
	}
	tx.Get(self, selectSQL, id)
	err = tx.Commit()
	return
}

// Get an record by ID.
func (self *Record) GetByID(db backfill.DB, id int) (err error) {
	sql := `SELECT * FROM records WHERE id = ? LIMIT 1;`
	err = db().Get(self, sql, id)
	return
}
