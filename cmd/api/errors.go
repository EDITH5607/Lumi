package main

import (
	"fmt"
	"net/http"
)


func (app *application) logError(r *http.Request, err error) {
	app.logger.Println(err)
}

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