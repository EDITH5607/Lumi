package main

import (
	"Green/internal/data"
	"encoding/json"
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
		Runtime int32	`json:"runtime"`
		Genres []string 	`json:"genres"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err!= nil {
		// we use err.Error() because err is a interface we need to call the function to get the err string
		app.errorResponse(w,r,http.StatusBadRequest, err.Error())
		return
	} 

	fmt.Fprintf(w,"%v\n",input)

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