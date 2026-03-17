package main

import (
	"Green/internal/data"
	"Green/internal/validator"
	"fmt"
	"net/http"
	"time"
)


func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status":"available",
		"system_info":map[string]string{
		"Environment":app.config.env,
		"version": version,
		},

	}

	err := app.writeJSON(data,w,http.StatusOK, nil) 
	if err != nil {
		app.serverErrorResponse(w,r,err)
		return
	}
	
}

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


	fmt.Fprintf(w,"%v\n",input)
	// app.writeJSON(envelope{"movie": input},w, http.StatusOK,  nil)

}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParams(r)
	if err!=nil {
		app.notFoundReponse(w,r)
		return
	}

	movie:= data.Movie{
		ID :		id,
		CreatedAt : time.Now(),
		Title : 	"Casablanca",
		Runtime :   102,
		Genres: []string{"drama", "romance", "war"},
		Version :1,
	}
	err = app.writeJSON(envelope{"movie":movie}, w, http.StatusOK, nil)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
		return
	}
}

func (app *application)updateMovieHandler(w http.ResponseWriter, r *http.Request){
	w.Write([]byte("Update movie Handler!!"))
}

func(app *application)deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("delete Movie Handler !!"))
}