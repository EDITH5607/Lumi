package main

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
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


func (app *application) rateLimit(next http.Handler) http.Handler {
	// 2 is the 2req per second (2 token refill in the bucket per second) 4 burst means max four entries in the bucket.
	limiter := rate.NewLimiter(2,4)


	// we wrap the inner anonymous function because http.handlerFunc will convert it to a http.Handler which is a interface satisfy servehttp method
	return  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() { // check the bucket is empty or not if empty then return false
			app.rateLimitExceededResponse(w,r)
			return
		}
		next.ServeHTTP(w,r)
	})
}