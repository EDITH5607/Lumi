package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("Record Not Found!!")
	ErrEditConflict = errors.New("Edit Conflict!!")
)

// struct for distributing db instances with all methods
type Models struct {
	Movies MovieModel // struct contain db instance for movies methods 
	Users  UserModel // struct contain db instance for user methods
}

// constructor to initialize the model struct
func NewModel(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{db: db},
		Users: UserModel{db: db},
	}
}

