package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"githup.com/ademdizdar/prime-api/internal/app"
)

func main() {
	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if len(os.Args) < 2 {
		log.Fatal("usage: migrate [up|down|status|redo|to <version>]")
	}
	if err := goose.Run(os.Args[1], db, "db/migrations", os.Args[2:]...); err != nil {
		log.Fatal(err)
	}
}
