package data

import (
	"errors"

	"github.com/jmoiron/sqlx"
)

var ErrRecordNotFound = errors.New("record not found")

type Models struct {
	Movies MovieModel
}

func NewModel(db *sqlx.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
