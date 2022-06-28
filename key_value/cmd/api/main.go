package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/abdulmajid18/keyVal/key_value/internal/data"
	"github.com/abdulmajid18/keyVal/key_value/internal/mailer"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Config struct {
	port    int
	env     string
	db      dbConfig
	smtp    smtp
	cors    cors
	limiter limiter
}
type smtp struct {
	host     string
	port     int
	username string
	password string
	sender   string
}
type application struct {
	config Config
	logger *log.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

type dbConfig struct {
	dsn          string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

// Add a cors struct and trustedOrigins field with the type []string.
type cors struct {
	trustedOrigins []string
}

// Add a new limiter struct containing fields for the requests-per-second and burst
// values, and a boolean field which we can use to enable/disable rate limiting
// altogether.
type limiter struct {
	rps     float64
	burst   int
	enabled bool
}

func getDataBaseEnvVariables() string {
	role := "ROLE"
	password := "PASSWORD"
	dbname := "DBNAME"

	currentWorkDirectory, _ := os.Getwd()
	file := currentWorkDirectory + "/cmd/api/.env"
	err := godotenv.Load(file)
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

const version = "1.0.0"

// Create a buildTime variable to hold the executable binary build time. Note that this
// must be a string type, as the -X linker flag will only work with string variables.
var buildTime string

func main() {
	var cfg Config

	// Publish a new "version" variable in the expvar handler containing our application
	// version number (currently the constant "1.0.0").
	expvar.NewString("version").Set(version)

	dsn := getDataBaseEnvVariables()
	flag.IntVar(&cfg.port, "port", 8000, "APi server port")
	flag.StringVar(&cfg.env, "env", "developmennt", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", dsn, "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as the default values. IMPORTANT: If you're following along,
	// make sure to replace the default values for smtp-username and smtp-password
	// with your own Mailtrap credentials.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "2457cf20620e26", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "e89e04c5e4baeb", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "RozzDB <no-reply@rozzdb.rozay.net>", "SMTP sender")

	// Use the flag.Func() function to process the -cors-trusted-origins command line
	// flag. In this we use the strings.Fields() function to split the flag value into a
	// slice based on whitespace characters and assign it to our config struct.
	// Importantly, if the -cors-trusted-origins flag is not present, contains the empty
	// string, or contains only whitespace, then strings.Fields() will return an empty
	// []string slice.
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	// Create command line flags to read the setting values into the config struct.
	// Notice that we use true as the default for the 'enabled' setting?
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg.db)
	if err != nil {
		log.
			Fatal(err)
	}

	// Publish the number of active goroutines.
	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))
	// Publish the database connection pool statistics.
	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))
	// Publish the current Unix timestamp.
	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	// Create a new version boolean flag with the default value of false.
	displayVersion := flag.Bool("version", false, "Display version and exit")
	flag.Parse()

	// If the version flag value is true, then print out the version number and
	// immediately exit.
	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		// Print out the contents of the buildTime variable.
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}
	err = app.serve()
	if err != nil {
		logger.Fatal(err)
	}

}
