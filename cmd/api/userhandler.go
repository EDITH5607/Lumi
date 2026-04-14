package main

import (
	"Green/internal/data"
	"Green/internal/validator"
	"errors"
	"net/http"
)


func (app *application) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

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

	err = user.Password.Set(input.Password)
	if err!=nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	v := validator.New()

	if data.ValidateUser(v,user); !v.Valid()	{
		app.failedValidationResponse(w,r,v.Errors)
		return 
	}

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

	err  = app.writeJSON(envelope{"user":user}, w, http.StatusCreated, nil)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
	}
}