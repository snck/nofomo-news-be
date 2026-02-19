package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() error {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		fmt.Println("DATABASE_URL environment variable is not set")
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	return DB.Ping()
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
