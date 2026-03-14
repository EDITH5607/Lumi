package data

import "time"


type Movie struct {
	ID int64			`json:"id"`
	CreatedAt time.Time	`json:"-"`	// "-" always remove that struct field from the json 
	Title string		`json:"title"`
	Year int32			`json:"year,omitempty"`
	Runtime Runtime		`json:"runtime,omitempty,string"` // this will convert the output of json as string rather than int
	Genres []string		`json:"genres,omitempty"` // omitempty only works if the struct field doesnt have any value or value is "", false, 0
	Version int32		`json:"version"`
}