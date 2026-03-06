package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

)

const version = "1.0.0"

type config struct {
	port int
	env string
}

type application struct {
	config config
	logger *log.Logger

}

func main() {

	var cfg config
	flag.IntVar(&cfg.port,"Port", 4000,"API Server Port")
	flag.StringVar(&cfg.env, "env", "development","Environment (development|staging|production)")
	flag.Parse()	

	logger := log.New(os.Stdout, "INFO: ", log.Ldate | log.Ltime)
	app := &application{
		config: cfg,
		logger: logger,
	}



	server :=  &http.Server{
		Addr: fmt.Sprintf(":%d",cfg.port),
		// the mux is httprouter.Route struct but it contain serveHTTP method to satisfy the Handler interface
		Handler: app.routes()	,
		ReadTimeout : 10*time.Second,
		WriteTimeout: 20* time.Second,
		IdleTimeout: time.Minute,
	}

	logger.Printf("Start %s Server on port: %d ",cfg.env,cfg.port)
	err := server.ListenAndServe()
	logger.Fatal(err)
	
}