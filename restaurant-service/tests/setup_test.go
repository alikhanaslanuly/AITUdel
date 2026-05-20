package tests

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func setupTestDB() *sql.DB {

	connStr := `
		host=localhost
		port=5432
		user=postgres
		password=369Zaq57
		dbname=restaurant_db
		sslmode=disable
	`

	db, err := sql.Open(
		"postgres",
		connStr,
	)

	if err != nil {
		log.Fatal(err)
	}

	return db
}
