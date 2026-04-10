package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

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
	// limiter := rate.NewLimiter(2,4)


	type client struct {
		limiter *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu sync.Mutex
		clients = make(map[string]*client)
	)


	// initializing the goroutine to act as a infinite loop which clean the client from map in every 1 min 
	// the condition for cleaning is if the client is not appeared > 3 minute
	go func ()  {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip,client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients,ip)
				}
			}
			mu.Unlock()
		}	
	}()


	

	// we wrap the inner anonymous function because http.handlerFunc will convert it to a http.Handler which is a interface satisfy servehttp method
	return  http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// splithostport used to split the host:port and return host,port,err
		ip,_, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w,r,err)
			return 
		}

		mu.Lock() // mutex lock is used to prevent the overwriting or race condition in map

		// providing rate limit for new clients
		if _,found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(2,4)}
		}

		//update the lastseen time to latest time
		clients[ip].lastSeen = time.Now()

		// calling Allow fn will consume one token if the bucket is empty  then it will return false
		if !clients[ip].limiter.Allow() { 
			mu.Unlock()
			app.rateLimitExceededResponse(w,r)
			return
		}

		// unlock the mutex which we used to prevent the race condition in map[ip]=ratelimit
		mu.Unlock() 
		next.ServeHTTP(w,r)
	})
}