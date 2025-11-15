package repository

import (
	"database/sql"
	"errors"

	"github.com/dmnAlex/shortener/internal/model/errx"
	"github.com/dmnAlex/shortener/internal/storage/pg"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type postgresRepo struct {
	db *pg.DB
}

func NewPostgresRepo(db *pg.DB) (URLRepository, error) {
	return &postgresRepo{db: db}, nil
}

const saveSQL = `
	INSERT INTO urls (short_id, original_url)
	VALUES (@short_id, @original_url)
`

func (r *postgresRepo) Save(shortID, url string) error {
	args := pgx.NamedArgs{
		"short_id":     shortID,
		"original_url": url,
	}

	res, err := r.db.Exec(saveSQL, args)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return errx.ErrAlreadyExists
	}

	return nil
}

const findSQL = `
	SELECT original_url
	FROM urls
	WHERE short_id = @short_id
`

func (r *postgresRepo) Find(shortID string) (string, error) {
	var originalURL string
	args := pgx.NamedArgs{
		"short_id": shortID,
	}

	err := r.db.QueryRow(findSQL, args, &originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errx.ErrNotFound
		}

		return "", err
	}

	return originalURL, nil
}

func (r *postgresRepo) Ping() error {
	return r.db.Ping()
}
