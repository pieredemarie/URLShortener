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

func (r *SQLiteRepo) CreateShortLink(ctx context.Context, shortLink, longURL string) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT into urls (short_code, long_url) VALUES (?, ?)",
		shortLink, longURL,
	)

	return err
}

func (r *SQLiteRepo) GetLongLink(ctx context.Context, shortUrl string) (string, error) {
	var longURL string

	err := r.db.QueryRowContext(ctx,
		"SELECT long_url FROM urls WHERE short_code = ?", shortUrl).Scan(&longURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("not found")
	}

	return longURL, err
}

func (r *SQLiteRepo) GetOrCreate(ctx context.Context, shortCode, longURL string) (string, error) {
	_, err := r.db.ExecContext(ctx,
		"INSERT OR IGNORE INTO urls (short_code, long_url) VALUES (?, ?)",
		shortCode, longURL,
	)
	if err != nil {
		return "", err
	}

	var existingCode string
	err = r.db.QueryRowContext(ctx,
		"SELECT short_code FROM urls WHERE long_url = ?",
		longURL,
	).Scan(&existingCode)

	return existingCode, err
}
