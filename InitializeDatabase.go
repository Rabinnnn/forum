package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func InitializeDB()(*sql.DB, error){
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil{
		return nil, err
	}

	_, err = db.Exec(`
	
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY NOT NULL,
			username TEXT UNIQUE,
			password TEXT,
			email TEXT UNIQUE
		);
		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`) 
	if err != nil{
		return nil, fmt.Errorf("error creating users table: %v", err)
	}	
	return db, nil
}