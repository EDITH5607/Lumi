package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("Record Not Found")
)

// struct for distributing db instances with all methods
type Models struct {
	Movies MovieModel // struct contain db instance for movies methods 
}

// constructor to initialize the model struct
func NewModel(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{db: db},
	}
}

