package repository

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type postgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(dsn string) (URLRepository, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	repo := &postgresRepo{db: db}

	if err := repo.Ping(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *postgresRepo) Save(shortID, url string) error {
	return nil
}

func (r *postgresRepo) Find(shortID string) (string, error) {
	return "", nil
}

func (r *postgresRepo) Ping() error {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	return r.db.PingContext(ctx)
}
