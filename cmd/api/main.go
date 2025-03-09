package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"

	"github.com/jcarloasilo/golang-rest-template/internal/database"
	"github.com/jcarloasilo/golang-rest-template/internal/env"
	"github.com/jcarloasilo/golang-rest-template/internal/smtp"
	"github.com/jcarloasilo/golang-rest-template/internal/version"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lmittmann/tint"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

type config struct {
	baseURL   string
	httpPort  int
	basicAuth struct {
		username       string
		hashedPassword string
	}
	cookie struct {
		secretKey string
	}
	db struct {
		database string
		password string
		username string
		port     string
		host     string
		schema   string
	}
	jwt struct {
		secretKey string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		from     string
	}
}

type application struct {
	config config
	db     *database.Queries
	dbPool *pgxpool.Pool
	logger *slog.Logger
	mailer *smtp.Mailer
	wg     sync.WaitGroup
}

func run(logger *slog.Logger) error {
	var cfg config

	cfg.baseURL = env.GetString("BASE_URL", "http://localhost:8080")
	cfg.httpPort = env.GetInt("HTTP_PORT", 8080)

	cfg.basicAuth.username = env.GetString("BASIC_AUTH_USERNAME", "admin")
	cfg.basicAuth.hashedPassword = env.GetString("BASIC_AUTH_HASHED_PASSWORD", "$2a$10$jRb2qniNcoCyQM23T59RfeEQUbgdAXfR6S0scynmKfJa5Gj3arGJa")
	cfg.cookie.secretKey = env.GetString("COOKIE_SECRET_KEY", "daapb3ukst43vpjsxf67ehomnlulacr3")
	cfg.jwt.secretKey = env.GetString("JWT_SECRET_KEY", "2sbhpt3ckvj5i5urt727fmeugwud7i3r")

	cfg.db.database = env.GetString("DB_DATABASE", "db")
	cfg.db.password = env.GetString("DB_PASSWORD", "pass")
	cfg.db.username = env.GetString("DB_USERNAME", "user")
	cfg.db.port = env.GetString("DB_PORT", "5432")
	cfg.db.host = env.GetString("DB_HOST", "localhost")
	cfg.db.schema = env.GetString("DB_SCHEMA", "public")

	cfg.smtp.host = env.GetString("SMTP_HOST", "example.smtp.host")
	cfg.smtp.port = env.GetInt("SMTP_PORT", 25)
	cfg.smtp.username = env.GetString("SMTP_USERNAME", "example_username")
	cfg.smtp.password = env.GetString("SMTP_PASSWORD", "pa55word")
	cfg.smtp.from = env.GetString("SMTP_FROM", "Example Name <no_reply@example.org>")

	showVersion := flag.Bool("version", false, "display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", cfg.db.username, cfg.db.password, cfg.db.host, cfg.db.port, cfg.db.database, cfg.db.schema)
	dbPool, err := database.NewPool(connStr)
	if err != nil {
		return err
	}
	log.Println("database connection pool established")
	defer dbPool.Close()

	db := database.New(dbPool)

	mailer, err := smtp.NewMailer(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.from)
	if err != nil {
		return err
	}
	log.Println("mailer connection established")

	app := &application{
		config: cfg,
		db:     db,
		dbPool: dbPool,
		logger: logger,
		mailer: mailer,
	}

	return app.serveHTTP()
}
