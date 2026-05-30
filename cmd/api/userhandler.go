package main

import (
	"Green/internal/data"
	"Green/internal/validator"
	"errors"
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


	//sending welcome email to the client
	err = app.mailer.Send(user.Email,"user_welcome.html", user)
	if err != nil {
		app.serverErrorResponse(w,r,err)
		return
	}

	// write the contents to the user like user is created(not activated)
	err  = app.writeJSON(envelope{"user":user}, w, http.StatusCreated, nil)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
	}
}