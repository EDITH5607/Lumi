package data

import (
	"Green/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// struct for saving the details of the movie for storing and retriving from db
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


// insert data into the db , movies struct -> db
func (m *MovieModel)Insert(movie *Movie) error{
	query := `INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, version`

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	return m.db.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m *MovieModel) Get(id int64) (*Movie, error) {
	if id<1 {
		return nil, ErrRecordNotFound
	}
	query := `SELECT id, created_at, title, year, runtime, genres, version
			FROM movies  WHERE id = $1`
	var movie Movie
	err := m.db.QueryRow(query,id).Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
	if err!=nil {
		switch {
		case errors.Is(err,sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil,err
		}
	}
	return &movie,nil
}

func (m *MovieModel) Update(movie *Movie) error {
	query :=  `UPDATE movies
			SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
			WHERE id = $5 AND version = $6
			RETURNING version`
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres),movie.ID,movie.Version}
	err := m.db.QueryRow(query,args...).Scan(&movie.Version)
	if err!=nil {
		switch {
		case errors.Is(err,sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err

		}

	}
	
	return nil
}



func (m *MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `DELETE FROM movies WHERE id = $1`
	result, err := m.db.Exec(query,id)
	if err!=nil {
		return err
	}
	
	roweffected,err := result.RowsAffected()
	if err!=nil {
		return  err
	}
	if roweffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}



// LIMIT says how many to take, OFFSET says how many to skip.  
//LIMIT is the pagesize and OFFSET is the page no.like how much need to skip
func (m *MovieModel)GetAll(title string, genres []string, f *Filters)([]*Movie, Metadata, error) {
	//the empty string "" , and the
	// filter condition in the SQL query will evaluate to true and act like it has been ‘skipped’.
	// Likewise with the genres parameter.
	query := fmt.Sprintf(`
			SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version
			FROM movies
			WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
			AND (genres @> $2 OR $2 = '{}')
			ORDER BY %s %s, id ASC
			LIMIT $3 OFFSET $4`, f.sortColumn(), f.sortDirection())

			// @> : ‘contains’ this means contains in the pq array if yes then return true otherwise false
	ctx , cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{title, pq.Array(genres), f.limit(), f.offset()}
	rows, err := m.db.QueryContext(ctx, query,args... )
	if err!=nil {
		return nil,Metadata{}, err
	}

	defer rows.Close()
	movies := []*Movie{}
	totalrecord := 0 
	for rows.Next() {
		var movie Movie
		err :=rows.Scan(
			&totalrecord,
			&movie.ID,
			&movie.CreatedAt, 
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err!=nil {
			return nil,Metadata{}, err
		}
		movies = append(movies,&movie)

	}
	if err = rows.Err(); err!=nil {
		return nil,Metadata{},err
	}
	metadata:= calculateMetaData(totalrecord, f.Page, f.PageSize)
	return movies,metadata, nil
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
	v.Check(movie.Runtime>0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres","must be provided" )
	v.Check(len(movie.Genres)>=1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres)<=5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genre", "must not contain duplicate values")

}
