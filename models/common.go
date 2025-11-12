package models

import (
	"backend/db"
	"database/sql"
)

func WithTx(fn func(tx *sql.Tx) error) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	return fn(tx)
}
