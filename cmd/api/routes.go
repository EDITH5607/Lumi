package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)


// all routes are registered here
func (app *application) routes() http.Handler {
	router := httprouter.New()

	// we are registor our methods to the httprouter because router itself send the reponse as plain text 
	// so we replace that with our json error reponse methods
	router.NotFound = http.HandlerFunc(app.notFoundReponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	//health route for checking the health of the system
	router.HandlerFunc(http.MethodGet, "/v1/health", app.healthCheckHandler)

	//movie routes
	router.HandlerFunc(http.MethodGet, "/v1/movies",app.listMovies)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	router.HandlerFunc(http.MethodPatch,"/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)



	// user routes
	router.HandlerFunc(http.MethodPost, "/v1/users",app.RegisterUser)

	// this returns a type  *httprouter.Router but this struct contain 
	// a serverHTTP method to satisfy the http.Handler interface
	return app.recoverPanic(app.rateLimit(router))
}