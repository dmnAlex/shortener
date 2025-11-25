package repository

import (
	"database/sql"
	"errors"
	"fmt"

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

const urlsSaveSQL = `
	INSERT INTO urls (short_id, original_url)
	VALUES (@short_id, @original_url)
	ON CONFLICT (original_url) DO UPDATE
	SET original_url = @original_url
	RETURNING short_id
`

const usersSaveSQL = `
	INSERT INTO users (user_id, short_id)
	VALUES (@user_id, @short_id)
	ON CONFLICT (user_id, short_id) DO NOTHING
`

func (r *postgresRepo) Save(userID, url string) (string, error) {
	shortID, err := utils.GenerateShortID()
	if err != nil {
		return "", err
	}

	var savedShortID string
	if err := r.DoTx(func(rTx *postgresRepo) error {
		args := pgx.NamedArgs{
			"short_id":     shortID,
			"original_url": url,
		}
		if err := rTx.db.QueryRow(urlsSaveSQL, args, &savedShortID); err != nil {
			return err
		}

		args = pgx.NamedArgs{
			"user_id":  userID,
			"short_id": savedShortID,
		}
		if _, err := rTx.db.Exec(usersSaveSQL, args); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return "", err
	}

	if shortID != savedShortID {
		return savedShortID, errx.ErrConflict
	}

	return shortID, nil
}

func (r *postgresRepo) SaveBatch(userID string, batch []model.ShortenBatchRequest) ([]model.ShortenBatchResponse, error) {
	var res []model.ShortenBatchResponse
	b := &pgx.Batch{}
	for _, item := range batch {
		shortID, err := utils.GenerateShortID()
		if err != nil {
			return nil, err
		}

		args := pgx.NamedArgs{
			"short_id":     shortID,
			"original_url": item.OriginalURL,
		}

		b.Queue(urlsSaveSQL, args)
	}

	if err := r.DoTx(func(rTx *postgresRepo) error {
		urlsBatchResults, err := rTx.db.SendBatch(b)
		if err != nil {
			return fmt.Errorf("send first batch: %w", err)
		}

		b = &pgx.Batch{}
		for i := range batch {
			item := model.ShortenBatchResponse{CorrelationID: batch[i].CorrelationID}
			if err := urlsBatchResults.QueryRow().Scan(&item.ShortURL); err != nil {
				return fmt.Errorf("process first batch: %w", err)
			}
			res = append(res, item)

			args := pgx.NamedArgs{
				"user_id":  userID,
				"short_id": item.ShortURL,
			}

			b.Queue(usersSaveSQL, args)
		}
		urlsBatchResults.Close()

		usersBatchResults, err := rTx.db.SendBatch(b)
		if err != nil {
			return fmt.Errorf("send second batch: %w", err)
		}

		for range batch {
			if _, err := usersBatchResults.Exec(); err != nil {
				return fmt.Errorf("process second batch: %w", err)
			}
		}
		usersBatchResults.Close()

		return nil
	}); err != nil {
		return nil, err
	}

	return res, nil
}

const urlsFindSQL = `
	SELECT original_url
	FROM urls
	WHERE short_id = @short_id
`

func (r *postgresRepo) Find(shortID string) (string, error) {
	var originalURL string
	args := pgx.NamedArgs{
		"short_id": shortID,
	}

	err := r.db.QueryRow(urlsFindSQL, args, &originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errx.ErrNotFound
		}

		return "", err
	}

	return originalURL, nil
}

const urlsFindAllSQL = `
	SELECT url.short_id, url.original_url
	FROM users usr
	JOIN urls url
	ON url.short_id = usr.short_id
	WHERE usr.user_id = @user_id
`

func (r *postgresRepo) FindAll(userID string) ([]model.UserURLsResponse, error) {
	args := pgx.NamedArgs{"user_id": userID}
	return pg.QueryMany(r.db, urlsFindAllSQL, pg.IfaceListFunc[*model.UserURLsResponse](), args)
}

func (r *postgresRepo) Ping() error {
	return r.db.Ping()
}

func (r *postgresRepo) DoTx(f func(r *postgresRepo) error, opts ...*pgx.TxOptions) error {
	return r.db.DoTx(func(mainDb *pg.DB) error {
		return f(NewPostgresRepo(mainDb))
	}, opts...)
}
