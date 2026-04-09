package main

import (
	"fmt"
	"net/http"
)

func (app *application)recoverPanic(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// deffer function used because when a panic happens go unwind the stack so 
		// the bottom of the stack we have the defer fns so it execute last
		defer func() {
			// built in recover fn to check if any panic happends or not
			if err:=recover(); err!=nil {
				w.Header().Set("Connection", "Close")  //automatically close the connection between go http server and browser
				app.serverErrorResponse(w,r,fmt.Errorf("%s",err)) //recover return 'any' type so we typecast it to a error type 
			}
		}()
		next.ServeHTTP(w,r)
	})

}