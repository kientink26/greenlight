package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kientink26/greenlight/internal/data"
	"github.com/kientink26/greenlight/internal/mailer"
	_ "github.com/lib/pq"
)

type config struct {
	env  string
	port int
	db   struct {
		dsn string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config config
	logger *log.Logger
	models data.Models
	mailer mailer.Mailer
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "api server port")
	flag.StringVar(&cfg.env, "env", "development", "environment")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "postgresql dsn")
	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "smtp host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "smtp port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "", "smtp username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "", "smtp password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@example.com>", "smtp sender")
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Println("Database connection established!")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.port),
		Handler: app.routes(),
	}
	fmt.Printf("Starting %s server on port %d!\n", cfg.env, cfg.port)
	err = server.ListenAndServe()
	logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}
