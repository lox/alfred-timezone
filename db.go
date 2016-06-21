package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var createSql = `
CREATE TABLE IF NOT EXISTS timezone (
    id INT PRIMARY KEY,
    name varchar(255) NOT NULL,
    country varchar(255) NOT NULL,
    timezone varchar(255) NOT NULL,
    population bigint
)`

func openDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./sqlite.db")
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec(createSql); err != nil {
		return nil, err
	}

	return db, nil
}
