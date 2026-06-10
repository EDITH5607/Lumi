package main

import (
	"Green/internal/data"
	"Green/internal/validator"
	"errors"
	"net/http"
	"time"
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



	err = app.model.Permissions.AddForUser(user.ID, "movies:read")
	if err!=nil {
		app.serverErrorResponse(w,r,err)
		return
	}


	token,err := app.model.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
		return
	}	

	

	app.background(func() {

		data := map[string]any{
			"activationToken":token.Plaintext,
			"userID":user.ID,
		}
		//sending welcome email to the client
		err = app.mailer.Send(user.Email,"user_welcome.html", data)
		if err != nil {
			// we use app.logger.printerror instead of app.serverError
			// This is because by the time we encounter the errors, the client will probably
			// have already been sent a 202 Accepted response by our writeJSON() helper.
			app.logger.PrintError(err,nil)
		}
	})




	// write the contents to the user like user is created(not activated)
	// The status code indicates the request has been accepted for processing, but
	// the processing has not been completed.
	err  = app.writeJSON(envelope{"user":user}, w, http.StatusAccepted, nil)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
	}

	// if any variable in the app code base if modified by the goroutine the variable in the app code base will reflect it.
}


func(app *application) ActivateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w,r,&input)
	if err!=nil {
		app.badRequestResponse(w,r,err)
		return
	}

	v := validator.New()

	if data.ValidateTokenPlaintext(v,input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w,r,v.Errors)
		return
	}

	user,err := app.model.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err!=nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w,r,v.Errors)
		default:
			app.serverErrorResponse(w,r,err)
		}
		return
	}
	
	user.Activated = true

	err = app.model.Users.Update(user)
	if err!=nil {
		switch{
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w,r)
		default:
			app.serverErrorResponse(w,r,err)
		}
		return
	}


	err = app.model.Tokens.DeleteAllForUser(user.ID, data.ScopeActivation)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
		return
	}

	err = app.writeJSON(envelope{"user":user}, w, http.StatusOK, nil)
	if err!=nil {
		app.serverErrorResponse(w,r,err)
	}


}