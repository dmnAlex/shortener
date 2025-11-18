package repository

import (
	"database/sql"
	"errors"

	"github.com/dmnAlex/shortener/internal/model"
	"github.com/dmnAlex/shortener/internal/model/errx"
	"github.com/dmnAlex/shortener/internal/storage/pg"
	"github.com/dmnAlex/shortener/internal/utils"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type postgresRepo struct {
	db *pg.DB
}

func NewPostgresRepo(db *pg.DB) *postgresRepo {
	return &postgresRepo{db: db}
}

const saveSQL = `
	INSERT INTO urls (short_id, original_url)
	VALUES (@short_id, @original_url)
	ON CONFLICT (original_url) DO UPDATE
	SET original_url = @original_url
	RETURNING short_id
`

func (r *postgresRepo) Save(url string) (string, error) {
	shortID, err := utils.GenerateShortID()
	if err != nil {
		return "", err
	}

	args := pgx.NamedArgs{
		"short_id":     shortID,
		"original_url": url,
	}

	var savedShortID string
	if err := r.db.QueryRow(saveSQL, args, &savedShortID); err != nil {
		return "", err
	}

	if shortID != savedShortID {
		return savedShortID, errx.ErrConflict
	}

	return shortID, nil
}

func (r *postgresRepo) SaveBatch(batch []model.ShortenBatchRequest) ([]model.ShortenBatchResponse, error) {
	var res []model.ShortenBatchResponse
	if err := r.DoTx(func(rTx *postgresRepo) error {
		for _, item := range batch {
			shortID, err := rTx.Save(item.OriginalURL)
			if err != nil && !errors.Is(err, errx.ErrConflict) {
				return err
			}

			res = append(res, model.ShortenBatchResponse{CorrelationID: item.CorrelationID, ShortURL: shortID})
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return res, nil
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

func (r *postgresRepo) DoTx(f func(r *postgresRepo) error, opts ...*sql.TxOptions) error {
	return r.db.DoTx(func(mainDb *pg.DB) error {
		return f(NewPostgresRepo(mainDb))
	}, opts...)
}
