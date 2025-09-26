package repository

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
)

type Postgres struct {
	db    *sql.DB
	dbURL string
}

func NewPostgres(dsn string) (*Postgres, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}
	return &Postgres{
		db:    db,
		dbURL: dsn,
	}, nil
}

func (p *Postgres) Test() string {
	return "Test"
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id SERIAL PRIMARY KEY,
            login VARCHAR(255) NOT NULL UNIQUE,
            password_hash VARCHAR(255) NOT NULL,
            create_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
        )`,
		`CREATE TABLE IF NOT EXISTS secret_data (
            id SERIAL PRIMARY KEY,
            user_id INTEGER NOT NULL,
            type VARCHAR(255) NOT NULL,
            name VARCHAR(255) NOT NULL,
            metadata TEXT,
            data BYTEA NOT NULL,
            version INTEGER DEFAULT 1,
            create_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
            update_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
            FOREIGN KEY (user_id) REFERENCES users(id)
        )`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w\nQuery: %s", err, query)
		}
	}
	return nil
}
