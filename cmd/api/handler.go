package main

import (
	"Green/internal/data"
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
	fmt.Fprintln(w, "Create movie handler!!!")

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