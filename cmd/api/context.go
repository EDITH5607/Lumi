package main

import (
	"Green/internal/data"
	"context"
	"net/http"
)

// making a context key type
type contextKey string


//initializing a contextkey as user 
const userContextKey = contextKey("user")


// used to set user context in the request context
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {

	// adding usercontext with user (struct) as value and make a shallow copy of the request
	ctx := context.WithValue(r.Context(), userContextKey, user)
	// make a shallow copy of the request and return it
	return r.WithContext(ctx)
}

//get the user details from the request context
func (app *application) contextGetUser(r *http.Request) *data.User {
	//get the user details from the request context and type convert it to User struct 
	user ,ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in Request context")
	}

	// return the user struct
	return user
}
