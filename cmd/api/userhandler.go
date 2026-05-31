package main

import (
	"Green/internal/data"
	"Green/internal/validator"
	"errors"
	"fmt"
	"net/http"
)

// register user from the input json data
func (app *application) RegisterUser(w http.ResponseWriter, r *http.Request) {
	/* One doubt is why we use input struct we use because of security concern 
	   the user struct have activate part , password which is not directly assigned
	   if we use readJson it directly assign the values to the struct so we have no control over it
	   to take the Explicit control over what to accept!!
	*/


	// making input struct for accepting data
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// read user data from the input json
	err := app.readJSON(w,r, &input)
	if err!=nil {
		app.badRequestResponse(w,r,err)
		return
	}

	// its like the object initalization &struct{} to a variable so methods can access the data directly and mutate the data
	user := &data.User{
		Name: input.Name,
		Email: input.Email,
		Activated: false,
	}

	// input password to user struct
	err = user.Password.Set(input.Password)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}


	//validate the data from the user
	v := validator.New()

	if data.ValidateUser(v,user); !v.Valid()	{
		app.failedValidationResponse(w,r,v.Errors)
		return 
	}

	// insert data to the database with some validations
	err = app.model.Users.Insert(user)
	if err!=nil {
		switch  {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w,r,v.Errors)
		default:
			app.serverErrorResponse(w,r,err)		
		}
		return
	}


	go func ()  {
		// if a goroutine panics (crashes) it will take down entire program, even if main programs is fine
		// defer runs after the surrounding function ends — whether it ends normally or by panic.
		// we make it as a defer fn because a block of code is need to run its not a goroutine
		// err handles expected errors. But some errors are unexpected panics (like nil pointer, index out of range).
		//  recover() catches those panics that err would never catch.
		defer func() {
			if err:= recover(); err!=nil {
				app.logger.PrintError(fmt.Errorf("%s", err),nil)
			}
		}()
		//sending welcome email to the client
		err = app.mailer.Send(user.Email,"user_welcome.html", user)
		if err != nil {
			// we use app.logger.printerror instead of app.serverError
			// This is because by the time we encounter the errors, the client will probably
			// have already been sent a 202 Accepted response by our writeJSON() helper.
			app.logger.PrintError(err,nil)
			return
		}
	}()



	// write the contents to the user like user is created(not activated)
	// The status code indicates the request has been accepted for processing, but
	// the processing has not been completed.
	err  = app.writeJSON(envelope{"user":user}, w, http.StatusAccepted, nil)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
	}

	// if any variable in the app code base if modified by the goroutine the variable in the app code base will reflect it.
}