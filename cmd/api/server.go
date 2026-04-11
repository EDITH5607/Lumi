package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	server :=  &http.Server{
		Addr: fmt.Sprintf(":%d",app.config.port),
		// the mux is httprouter.Route struct but it contain serveHTTP method to satisfy the Handler interface
		Handler: app.routes()	,
		ReadTimeout : 10*time.Second,
		WriteTimeout: 20* time.Second,
		IdleTimeout: time.Minute,
	}

	// create shutdownError channel for recieving any error returned by the graceful shutdown	function
	shutdownError := make(chan error) // unbuffered channel so reciever need to recieve at the same time sending


	// consider this part as another parallel process which use shutdownerror channel to communicate with serve fn
	go func ()  {

		// making a channel with type osSignal(only recieve these type signal) and room size is 1
		quit := make(chan os.Signal, 1)
		//notify and store if these signals are awoked
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

		// Read the signal from quit channel(), if channel is empty rest of the code will never run(the goroutine block code not the other part of funcition)
		s := <-quit
		//logging the signal for grace shutdown
		app.logger.PrintInfo("Shutting down server", map[string]string{
			"signal":s.String(),
		})


		//setting context timeout of 20 seconds for graceful shutdown, max time to shutdown the application 
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()


		// Close your DB connections here!
		app.db.Close()
		app.logger.PrintInfo("DB close Successfully",nil)

		// returning error message when server shutdown to the shutdownerror channel
		// the sends nil or error to the serve fn using shutdownerror channel 
		// when complete shutdown the max limit is 20 second
		shutdownError <- server.Shutdown(ctx)

	}()


	// logging server started
	app.logger.PrintInfo("Start Server", map[string]string{
		"addr":server.Addr, 
		"env":app.config.env,
	})
	

	// Calling Shutdown() on our server will cause ListenAndServe() to immediately
	// return a http.ErrServerClosed error. So if we see this error, it is actually a
	// good thing and an indication that the graceful shutdown has started. So we check
	// specifically for this, only returning the error if it is NOT http.ErrServerClosed.
	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return  err
	}


	// recieves the nil or error message from the goroutine
	// this block the rest of the code until it recieve nil or error from goroutine, 
	// so app.logger.... part run until this recive the corresponding value 
	err = <-shutdownError
	if err!=nil {
		return err
	}

	app.logger.PrintInfo("Stopped Server",map[string]string{
		"addr":server.Addr,
	})

	return nil
}