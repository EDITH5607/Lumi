package main

import (
	"Green/internal/data"
	"Green/internal/jsonlog"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// _ in import used to o stop the Go compiler complaining that the package isn't being used.
const version = "1.0.0"

type config struct {
	port int
	env string
	db struct {
		dsn string	
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	model data.Models // contains the Models struct

}

func main() {

	var cfg config
	flag.IntVar(&cfg.port,"Port", 4000,"API Server Port")
	flag.StringVar(&cfg.env, "env", "development","Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("LUMI_DB_DSN"), "PostgresSQL DSN")
	flag.Parse()	

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err!=nil {
		logger.PrintFatal(err,nil)
	}
	defer db.Close()
	logger.PrintInfo("Database connection pool Established Successfully!!",nil)

	app := &application{
		config: cfg,
		logger: logger,
		model: data.NewModel(db), // return Model struct

	}



	server :=  &http.Server{
		Addr: fmt.Sprintf(":%d",cfg.port),
		// the mux is httprouter.Route struct but it contain serveHTTP method to satisfy the Handler interface
		Handler: app.routes()	,
		ReadTimeout : 10*time.Second,
		WriteTimeout: 20* time.Second,
		IdleTimeout: time.Minute,
	}

	logger.PrintInfo("Start Server", map[string]string{"addr":server.Addr, "env":cfg.env})
	err = server.ListenAndServe()
	logger.PrintFatal(err, nil)
	
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres",cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	// creating a root context(parent) and give it a signal like timeout when 5 seconds over..
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// cancel the context when the fn returns 
	defer cancel()

	// ping to the db with a fixed time so that we know db is working or not
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}