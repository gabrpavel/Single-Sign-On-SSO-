package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	config2 "sso/internal/config"
)

func main() {
	var configPath, migrationsPath, migrationsTable string
	config := config2.MustLoad()

	flag.StringVar(&configPath, "config-path", "", "path to config")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.Parse()

	if configPath == "" {
		panic("config-path is required")
	}
	if migrationsPath == "" {
		panic("migrationsPath is required")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&x-migrations-table=%s",
			config.Db.User, config.Db.Password, config.Db.Host, config.Db.Port, config.Db.DBName, migrationsTable),
	)

	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")

			return
		}

		panic(err)
	}

	fmt.Println("migrations applied successfully")
}
