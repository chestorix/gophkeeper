package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/chestorix/gophkeeper/internal/errors"
	"github.com/chestorix/gophkeeper/internal/models"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
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

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			login VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS secret_data (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			type VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			metadata TEXT,
			data BYTEA NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			UNIQUE(user_id, name)
		)`,

		`CREATE INDEX IF NOT EXISTS idx_secret_data_user_id ON secret_data(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_secret_data_updated_at ON secret_data(updated_at)`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to execute query: %w\nQuery: %s", err, query)
		}
	}
	return nil
}

func (p *Postgres) CreateUsers(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id, login, password_hash, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5)`

	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := p.db.ExecContext(ctx, query, user.ID, user.Login, user.PasswordHash,
		user.CreatedAt, user.UpdatedAt)
	return err
}
func (p *Postgres) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	query := `SELECT id,login,password_hash,created_at,updated_at FROM users WHERE login=$1`
	var user models.User
	err := p.db.QueryRowContext(ctx, query, login).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (p *Postgres) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `SELECT id,login,password_hash,created_at,updated_at FROM users WHERE id=$1`
	var user models.User
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.ErrUserNotFound
	}
	return &user, err
}
func (p *Postgres) SaveSecretData(ctx context.Context, data *models.SecretItemData) error {
	query := `INSERT INTO secret_data (id, user_id, type, name, metadata, data, version, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	data.ID = uuid.New().String()
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()

	_, err := p.db.ExecContext(ctx, query,
		data.ID, data.UserID, string(data.Type), data.Name, data.Metadata,
		data.Data, data.Version, data.CreatedAt, data.UpdatedAt)
	return err
}

func (p *Postgres) GetSecretDataByID(ctx context.Context, id string, userID string) (*models.SecretItemData, error) {
	query := `SELECT id, user_id, type, name, metadata, data, version, created_at, updated_at 
              FROM secret_data WHERE id=$1 AND user_id=$2`

	var data models.SecretItemData
	var typeStr string

	err := p.db.QueryRowContext(ctx, query, id, userID).Scan(
		&data.ID, &data.UserID, &typeStr, &data.Name, &data.Metadata,
		&data.Data, &data.Version, &data.CreatedAt, &data.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.ErrDataNotFound
	}
	if err != nil {
		return nil, err
	}

	data.Type = models.DataType(typeStr)
	return &data, nil
}

func (p *Postgres) GetUserSecretData(ctx context.Context, userID string, lastSync time.Time) ([]models.SecretItemData, error) {
	query := `SELECT id, user_id, type, name, metadata, data, version, created_at, updated_at
              FROM secret_data WHERE user_id = $1 AND updated_at > $2
              ORDER BY updated_at DESC`

	rows, err := p.db.QueryContext(ctx, query, userID, lastSync)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []models.SecretItemData
	for rows.Next() {
		var item models.SecretItemData
		var typeStr string

		err := rows.Scan(
			&item.ID, &item.UserID, &typeStr, &item.Name, &item.Metadata,
			&item.Data, &item.Version, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		item.Type = models.DataType(typeStr)
		data = append(data, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *Postgres) UpdateSecretData(ctx context.Context, data *models.SecretItemData) error {
	query := `UPDATE secret_data 
              SET type = $1, name = $2, metadata = $3, data = $4, version = version + 1, updated_at = $5
              WHERE id = $6 AND user_id = $7`

	data.UpdatedAt = time.Now()
	result, err := p.db.ExecContext(ctx, query,
		string(data.Type), data.Name, data.Metadata, data.Data, data.UpdatedAt,
		data.ID, data.UserID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.ErrDataNotFound
	}
	return nil
}

func (p *Postgres) DeleteSecretData(ctx context.Context, id, userID string) error {
	query := `DELETE FROM secret_data WHERE id = $1 AND user_id = $2`
	result, err := p.db.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.ErrDataNotFound
	}
	return nil
}
func (p *Postgres) Ping(ctx context.Context) error {
	return p.db.PingContext(ctx)
}

func (p *Postgres) GetSecretDataByName(ctx context.Context, name, userID string) (*models.SecretItemData, error) {
	query := `SELECT id, user_id, type, name, metadata, data, version, created_at, updated_at 
              FROM secret_data WHERE name=$1 AND user_id=$2`

	var data models.SecretItemData
	var typeStr string

	err := p.db.QueryRowContext(ctx, query, name, userID).Scan(
		&data.ID, &data.UserID, &typeStr, &data.Name, &data.Metadata,
		&data.Data, &data.Version, &data.CreatedAt, &data.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, errors.ErrDataNotFound
	}
	if err != nil {
		return nil, err
	}

	data.Type = models.DataType(typeStr)
	return &data, nil
}

func (p *Postgres) DeleteSecretDataByName(ctx context.Context, name, userID string) error {
	query := `DELETE FROM secret_data WHERE name = $1 AND user_id = $2`
	result, err := p.db.ExecContext(ctx, query, name, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.ErrDataNotFound
	}
	return nil
}
