package repository

import (
	"context"
	"database/sql"
	"errors"
)

type SQLiteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(dbPath string) (*SQLiteRepo, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			short_code VARCHAR(10) UNIQUE NOT NULL,
			long_url TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			clicks INTEGER DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
	`)

	if err != nil {
		return nil, err
	}

	return &SQLiteRepo{db: db}, nil
}

func (r *SQLiteRepo) Create(ctx context.Context, longURL string) (uint64, error) {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO urls (long_url, short_code) VALUES (?, '')",
		longURL,
	)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

func (r *SQLiteRepo) UpdateShortCode(
	ctx context.Context,
	id uint64,
	shortCode string,
) error {

	_, err := r.db.ExecContext(ctx,
		"UPDATE urls SET short_code = ? WHERE id = ?",
		shortCode,
		id,
	)

	return err
}

func (r *SQLiteRepo) GetLongLink(
	ctx context.Context,
	shortCode string,
) (string, error) {

	var longURL string

	err := r.db.QueryRowContext(ctx,
		"SELECT long_url FROM urls WHERE short_code = ?",
		shortCode,
	).Scan(&longURL)

	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("not found")
	}

	return longURL, err
}

func (r *SQLiteRepo) GetShortLink(
	ctx context.Context,
	longURL string,
) (string, error) {

	var shortCode string

	err := r.db.QueryRowContext(ctx,
		"SELECT short_code FROM urls WHERE long_url = ?",
		longURL,
	).Scan(&shortCode)

	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("not found")
	}

	return shortCode, err
}
