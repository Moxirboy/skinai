package postgres

import "database/sql"

type TipsRepository struct {
	db *sql.DB
}

func NewTipsRepository(db *sql.DB) *TipsRepository {
	return &TipsRepository{
		db: db,
	}
}
