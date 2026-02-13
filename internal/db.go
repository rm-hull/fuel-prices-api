package internal

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"strings"
)

//go:embed sql/migration.sql
var migrationSQL string


// //go:embed sql/insert_company_data.sql
// var InsertCompanyDataSQL string

// //go:embed sql/search.sql
// var SearchSQL string

func CreateDB(db *sql.DB) error {
	_, err := db.Exec(migrationSQL)
	return err
}

func Connect(dbPath string) (*sql.DB, error) {
	dsn := dbPath
	if strings.Contains(dsn, "?") {
		dsn += "&"
	} else {
		dsn += "?"
	}
	queryParams := []string{"_busy_timeout=5000", "_journal_mode=WAL"}
	dsn += strings.Join(queryParams, "&")
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Printf("connected to database: %s", dsn)

	err = CreateDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}
	return db, nil
}

