package main

import (
	"Green/internal/data"
	"Green/internal/validator"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"
)



// used to check the health of api
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status":"available",
		"system_info":map[string]string{
		"Environment":app.config.env,
		"version": version,
		},

	}
	time.Sleep(4*time.Second)
	err := app.writeJSON(data,w,http.StatusOK, nil) 
	if err != nil {
		app.serverErrorResponse(w,r,err)
		return
	}
	
}


// create movie in the route
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Title string `json:"title"`
		Year int32	`json:"year"`
		Runtime data.Runtime	`json:"runtime"`
		Genres []string 	`json:"genres"`
	}
	err := app.readJSON(w, r, &input)
	if err!= nil {
		app.badRequestResponse(w,r,err)
		return
	} 

	// we make it as a pointer because te validatemovie function need pointer as input
	movie := &data.Movie{
		Title: input.Title,
		Year: input.Year,
		Runtime: input.Runtime,
		Genres: input.Genres,
	}
	v := validator.New()
	
	if data.ValidateMovie(v, movie); !v.Valid()  { // if the len(v.Error) is not empty
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.model.Movies.Insert(movie)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("v1/movies/%d", movie.ID))

	// fmt.Fprintf(w,"%v\n",input)
	err = app.writeJSON(envelope{"movie": input},w, http.StatusCreated,  headers)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
	}

}


// show movie according the give id
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParams(r)
	if err!=nil {
		app.notFoundReponse(w,r)
		return
	}

	movie,err := app.model.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err,data.ErrRecordNotFound):
			app.notFoundReponse(w,r)
		default:
			app.serverErrorResponse(w,r,err)
		}
		return 
	}
	err = app.writeJSON(envelope{"movie":movie}, w, http.StatusOK, nil)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
		return
	}
}


// update the movie on the db
func (app *application)updateMovieHandler(w http.ResponseWriter, r *http.Request){
	id, err := app.readIDParams(r)
	if err!= nil {
		app.notFoundReponse(w,r)
	}
	movie,err := app.model.Movies.Get(id) // fetch from db to check the id of the movie is valid or not
	if err != nil {
		switch {
		case errors.Is(err,sql.ErrNoRows):
			app.notFoundReponse(w,r)
		default:
			app.serverErrorResponse(w,r,err)			
		}
	}


	//reading content from json
	// When you decode a JSON request body into a Go struct,
	//  any fields not included in the JSON automatically get their zero-value: like int->0, string->"", bool->false ... pointer->nil
	/*
	This creates an ambiguity problem — you can't tell the difference between:
	A client sending {"title": ""} → meaning they intentionally provided an empty value (should trigger a validation error)
	A client not sending title at all → meaning they want to skip updating that field (should be silently ignored)
	Since pointers have a zero-value of nil, you can change your struct fields to use pointer types instead of plain types.
	Then the logic becomes simple:
	Field is nil → client didn't send it → skip it
	Field is not nil → client sent a value (even if empty) → validate and update it
	*/
	// var input struct {
	// 	Title string `json:"title"`
	// 	Year int32 `json:"year"`
	// 	Runtime data.Runtime `json:"runtime"`
	// 	Genres []string `json:"genres"`
	// }

	var input struct {
		Title *string `json:"title"`
		Year *int32 `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres []string `json:"genres"`
	}
	err = app.readJSON(w,r,&input)
	if err !=nil {
		app.badRequestResponse(w,r,err)
		return
	}

	if input.Title!=nil {
		movie.Title = *input.Title
	}
	if input.Year!=nil {
		movie.Year = *input.Year
	}
	if input.Runtime!=nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres!=nil {
		movie.Genres = input.Genres
	}

	v := validator.New()
	if data.ValidateMovie(v,movie); !v.Valid() {
		app.failedValidationResponse(w,r,v.Errors)
		return
	}
	
	err = app.model.Movies.Update(movie)
	if err != nil {
		switch  {
		case errors.Is(err,data.ErrEditConflict):
			app.editConflictResponse(w,r)
		default:
			app.serverErrorResponse(w,r,err)
		}
		return
	}

	err = app.writeJSON(envelope{"movie":movie}, w,http.StatusOK,nil)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
	}
}


// delete movie from the database using this route
func(app *application)deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id,err := app.readIDParams(r)
	if err!=nil{
		app.badRequestResponse(w,r,err)
		return
	}

	err = app.model.Movies.Delete(id)
	if err!=nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundReponse(w,r)
		default:
			app.serverErrorResponse(w,r,err)
		}
		return
	}
	err = app.writeJSON(envelope{"message":"Movie Successfully deleted"},w,http.StatusOK,nil)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
	}


}


// list all movies like pagination and filter
func (app *application) listMovies(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string
		Genres []string
		data.Filters
	}
	v := validator.New()
	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs,"genres",[]string{})
	input.Filters.Page = app.readInt(qs, "page",1,v)
	input.Filters.PageSize = app.readInt(qs, "page_size",20,v)
	input.Filters.Sort = app.readString(qs, "sort","id")
	input.Filters.SafeSortlist = []string{"id", "title", "runtime", "-id", "-title", "-runtime"}
	if data.ValidateFilters(v,&input.Filters); !v.Valid() {
		app.failedValidationResponse(w,r,v.Errors)
		return
	}

	movies,metadata,  err := app.model.Movies.GetAll(input.Title, input.Genres, &input.Filters)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
		return
	}

	err = app.writeJSON(envelope{"metadata":metadata,"movies":movies}, w,http.StatusOK, nil)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
		
	}

	// fmt.Fprintf(w,"%+v", input) // +v gives the struct with its member name like title:hello,...
	//%v only gives the value


}