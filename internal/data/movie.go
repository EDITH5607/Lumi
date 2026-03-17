package data

import (
	"Green/internal/validator"
	"database/sql"
	"time"
)


type Movie struct {
	ID int64			`json:"id"`
	CreatedAt time.Time	`json:"-"`	// "-" always remove that struct field from the json 
	Title string		`json:"title"`
	Year int32			`json:"year,omitempty"`
	Runtime Runtime		`json:"runtime,omitempty,string"` // this will convert the output of json as string rather than int
	Genres []string		`json:"genres,omitempty"` // omitempty only works if the struct field doesnt have any value or value is "", false, 0
	Version int32		`json:"version"`
}


// struct contain the db instance and methods 
type MovieModel struct {
	db *sql.DB
}

func (m *MovieModel)Insert(movie *Movie) error{
	return nil
}

func (m *MovieModel) Get(id int64) (*Movie, error) {
	return nil,nil
}

func (m *MovieModel) Update(movie *Movie) error {
	return nil
}

func (m *MovieModel) Delete(id int64) error {
	return nil
}


func ValidateMovie(v *validator.Validator, movie *Movie) {
	//If that statement is True, the validator stays quiet. -> this is how the check function works
	// If that statement is False, the validator says, "Wait,
	//  that's not okay!" and adds the error message.
	v.Check(movie.Title != "", "title", "must be provided!!")
	v.Check(len(movie.Title)<500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year!=0, "year", "must be provided!!")
	v.Check(movie.Year>1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()),"year", "must not be in the future")
	
	v.Check(movie.Runtime!=0, "runtime", "must be provided!!")
	v.Check(movie.Runtime<0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres","must be provided" )
	v.Check(len(movie.Genres)>=1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres)<=5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genre", "must not contain duplicate values")

}
