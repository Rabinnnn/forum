package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var Globaldb *sql.DB

func InitializeDB() (*sql.DB, error) {
	var err error
	Globaldb, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		return nil, err
	}

	// Create all tables in a single transaction
	tx, err := Globaldb.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// SQL statements for creating tables
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY NOT NULL,
			username TEXT UNIQUE,
			password TEXT,
			email TEXT UNIQUE,
			profile_pic TEXT
		);
		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,

		`CREATE TABLE IF NOT EXISTS posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT,
			title TEXT NOT NULL,
			content TEXT,
			image_path TEXT,
			likes INTEGER DEFAULT 0,
			dislikes INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);`,

		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS post_categories (
			post_id TEXT NOT NULL,
			category_id INTEGER NOT NULL,
			PRIMARY KEY (post_id, category_id),
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (category_id) REFERENCES categories(id)
		);`,

		`CREATE TABLE IF NOT EXISTS comments (
			id TEXT PRIMARY KEY NOT NULL,
			post_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);`,

		`CREATE TABLE IF NOT EXISTS likes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			post_id TEXT NOT NULL,
			like INTEGER NOT NULL CHECK (like IN (0, 1)), -- 1 for like, 0 for dislike
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id),
			FOREIGN KEY (post_id) REFERENCES posts(id),
			UNIQUE(user_id, post_id)
		);
		CREATE INDEX IF NOT EXISTS idx_likes_post_id ON likes(post_id);`,
	}

	// Execute each table creation statement
	for _, table := range tables {
		if _, err := tx.Exec(table); err != nil {
			return nil, fmt.Errorf("error creating table: %v", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %v", err)
	}

	return Globaldb, nil
}
