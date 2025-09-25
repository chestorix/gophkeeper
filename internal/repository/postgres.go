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

	//if err := createTables(db); err != nil {
	//	return nil, fmt.Errorf("failed to create tables: %w", err)
	//}
	return &Postgres{
		db:    db,
		dbURL: dsn,
	}, nil
}

func (p *Postgres) Test() string {
	return "Test"
}
