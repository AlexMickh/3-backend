package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	username := os.Getenv("DB_USER")
	if username == "" {
		log.Fatal("DB_USER is required")
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		log.Fatal("DB_PASSWORD is required")
	}
	host := os.Getenv("DB_HOST")
	if host == "" {
		log.Fatal("DB_HOST is required")
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		log.Fatal("DB_PORT is required")
	}
	database := os.Getenv("DB_NAME")
	if database == "" {
		log.Fatal("DB_NAME is required")
	}
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if database == "" {
		log.Fatal("MIGRATIONS_PATH is required")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			username,
			password,
			host,
			port,
			database,
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("no migrations to apply")
			return
		}
		log.Fatal(err)
	}

	log.Println("migrations applied")
}
