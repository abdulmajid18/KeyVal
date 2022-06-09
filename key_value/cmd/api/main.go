package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/abdulmajid18/keyVal/key_value/internal/data"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const version = "1.0"

type Config struct {
	port int
	env  string
	db   dbConfig
}

type application struct {
	config Config
	logger *log.Logger
	models data.Models
}

type dbConfig struct {
	dsn          string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func getDataBaseEnvVariables() string {
	role := "ROLE"
	password := "PASSWORD"
	dbname := "DBNAME"

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("An error occured loading .env")
	}
	role = os.Getenv(role)
	password = os.Getenv(password)
	dbname = os.Getenv(dbname)

	return fmt.Sprintf("postgres://%s:%s@localhost/%s", role, password, dbname)

}

func openDB(dbConfig dbConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbConfig.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(dbConfig.maxOpenConns)
	db.SetMaxIdleConns(dbConfig.maxIdleConns)

	duraion, err := time.ParseDuration(dbConfig.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duraion)

	// Create a context with a 5-second timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Use PingContext() to establish a new connection to the database, passing in the
	// context we created above as a parameter. If the connection couldn't be
	// established successfully within the 5 second deadline, then this will return an
	// error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, err
}

func main() {
	var cfg Config

	dsn := getDataBaseEnvVariables()
	flag.IntVar(&cfg.port, "port", 8000, "APi server port")
	flag.StringVar(&cfg.env, "env", "developmennt", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", dsn, "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg.db)
	if err != nil {
		log.
			Fatal(err)
	}
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}
	fmt.Println("Starting Server!")
	routes := app.routes()
	server := &http.Server{
		Handler: routes,
		Addr:    "127.0.0.1:8000",
	}
	logger.Printf("starting %s server on %s", cfg.env, server.Addr)
	err = server.ListenAndServe()
	logger.Fatal(err)

}
