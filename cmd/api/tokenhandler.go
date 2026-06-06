package main

import (
	"Green/internal/data"
	"Green/internal/validator"
	"errors"
	"net/http"
	"time"
)

func(app *application) createAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	var input struct{
		Email string 	`json:"email"`
		Password string	`json:"password"`
	}

	err := app.readJSON(w,r,&input)
	if err!=nil {
		app.badRequestResponse(w,r,err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v,input.Email)
	data.ValidatePassword(v,input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w,r,v.Errors)
		return
	}

	user, err := app.model.Users.GetByEmail(input.Email)
	if err!=nil {
		switch {
		case errors.Is(err,data.ErrRecordNotFound):
			// app.invalid
		default:
			app.serverErrorResponse(w,r,err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
		return
	}

	if !match {
		// app.invali
		return
	}

	token, err := app.model.Tokens.New(user.ID,24*time.Hour, data.ScopeAuthentication)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
		return
	}

	err = app.writeJSON(envelope{"authentication_token":token}, w,http.StatusCreated, nil)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
	}


}