package main

import (
	"fmt"
	"net/http"
)

// used to log error to the terminal
func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method":r.Method,
		"request_url": r.URL.String(),
	})
}


// pass the error as json to the client
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request,status int,  message any) {
	env := envelope{"error":message}
	err := app.writeJSON(env,w,status, nil)
	if err!= nil {
		app.logError(r,err)
		w.WriteHeader(http.StatusInternalServerError) 
	}
}


// error is an interface if you pass the err to the fmt.log,fmt.print it will automatically 
// call the interface funtion inside the object string() or Error() , 
// if we need to pass the string of the error we need to call the err.Error() method
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w,r,http.StatusInternalServerError, message)
}

func (app *application) notFoundReponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}


func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w,r,http.StatusMethodNotAllowed, message)
}

func (app *application)failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w,r,http.StatusUnprocessableEntity, errors)
}

func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit Conflict, please try again"
	app.errorResponse(w,r,http.StatusConflict,message)
}


func (app *application)rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w,r,http.StatusTooManyRequests, message)
}

func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(w,r,http.StatusUnauthorized, message)
}


func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {

	// help inform or
	// remind the client that we expect them to authenticate using a bearer token
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "Invalid or missing authentication token"
	app.errorResponse(w,r,http.StatusUnauthorized, message)
}



func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errorResponse(w,r,http.StatusUnauthorized, message)
}

func (app *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"

	app.errorResponse(w,r,http.StatusForbidden, message) 
}


func (app *application) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account doesn't have the necessary permissions to access this resource"
	app.errorResponse(w,r,http.StatusForbidden, message)
}


