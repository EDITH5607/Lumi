package main

import (
	"Green/internal/data"
	"Green/internal/jsonlog"
	"Green/internal/mailer"
	"context"
	"database/sql"
	"flag"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// _ in import used to o stop the Go compiler complaining that the package isn't being used.
const version = "1.0.0"

type config struct {
	port int
	env  string
	db struct {
		dsn string	
	}
	limiter struct {
		rps     float64 //request per second usual the ratelimiter internally use float for this for calculation
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	model data.Models // contains the Models struct
	db *sql.DB
	mailer mailer.Mailer

}

func main() {

	var cfg config

	// Filling config struct
	// port, stage,dsn(data source name) for db
	flag.IntVar(&cfg.port,"Port", 4000,"API Server Port")
	flag.StringVar(&cfg.env, "env", "development","Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("LUMI_DB_DSN"), "PostgresSQL DSN")


	// ratelimiter cli arguments
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2,"Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst,"limiter-burst",4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")


	//smtp server configeration cli arguments
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "myname", "SMTP-username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "your password", "SMTP-password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Lumi <no-reply@lumi.myemail.net>", "SMTP-sender")
	
	flag.Parse()	




	// making logger instance
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	// making db instance
	db, err := openDB(cfg)
	if err!=nil {
		logger.PrintFatal(err,nil)
	}
	defer db.Close()
	logger.PrintInfo("Database connection pool Established Successfully!!",nil)

	
	// filling application struct 
	app := &application{  
		config: cfg,  // config struct is passed
		logger: logger,	//logger instance is passed
		model: data.NewModel(db), // return Model struct which stores all model for dependecy injection
		db: db,  // filling db instance here pg instance is passed
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender), // providing new instance of mail service for all handlers

	}


	// Starting the server using custom serve function
	err = app.serve()
	if err!=nil {
		logger.PrintFatal(err, nil)
	}
	
}


// Returning db instance according to the dsn
func openDB(cfg config) (*sql.DB, error) {
	// open database
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