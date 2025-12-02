package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dmnAlex/shortener/internal/logger"
	"github.com/dmnAlex/shortener/internal/model"
	"github.com/dmnAlex/shortener/internal/model/errx"
	"github.com/dmnAlex/shortener/internal/storage/pg"
	"github.com/dmnAlex/shortener/internal/utils"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

const (
	deleteChSize  = 1000
	batchSize     = 100
	flushInterval = 1 * time.Second
)

type postgresRepo struct {
	db       *pg.DB
	deleteCh chan model.DeleteTask
	wg       sync.WaitGroup
}

func NewPostgresRepo(db *pg.DB) *postgresRepo {
	r := &postgresRepo{
		db:       db,
		deleteCh: make(chan model.DeleteTask, deleteChSize),
	}

	r.wg.Add(1)
	go r.deletionWorker()

	return r
}

const saveURLSQL = `
	INSERT INTO urls (short_id, original_url, user_id)
	VALUES (@short_id, @original_url, @user_id)
	ON CONFLICT (original_url) DO UPDATE
	SET original_url = @original_url
	RETURNING short_id
`

func (r *postgresRepo) Save(userID, url string) (string, error) {
	shortID, err := utils.GenerateShortID()
	if err != nil {
		return "", err
	}

	args := pgx.NamedArgs{
		"short_id":     shortID,
		"original_url": url,
		"user_id":      userID,
	}

	var savedShortID string
	if err := r.db.QueryRow(saveURLSQL, args, &savedShortID); err != nil {
		return "", nil
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
			"user_id":      userID,
		}

		b.Queue(saveURLSQL, args)
	}

	urlsBatchResults, err := r.db.SendBatch(b)
	if err != nil {
		return nil, fmt.Errorf("send first batch: %w", err)
	}

	b = &pgx.Batch{}
	for i := range batch {
		item := model.ShortenBatchResponse{CorrelationID: batch[i].CorrelationID}
		if err := urlsBatchResults.QueryRow().Scan(&item.ShortURL); err != nil {
			return nil, fmt.Errorf("process first batch: %w", err)
		}
		res = append(res, item)
	}
	urlsBatchResults.Close()

	return res, nil
}

const findURLSQL = `
	SELECT original_url, is_deleted
	FROM urls
	WHERE short_id = @short_id
`

func (r *postgresRepo) Find(shortID string) (string, error) {
	var originalURL string
	var isDeleted bool

	args := pgx.NamedArgs{
		"short_id": shortID,
	}

	err := r.db.QueryRow(findURLSQL, args, &originalURL, &isDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errx.ErrNotFound
		}

		return "", err
	}

	if isDeleted {
		return "", errx.ErrGone
	}

	return originalURL, nil
}

const findAllURLsSQL = `
	SELECT short_id, original_url
	FROM urls
	WHERE user_id = @user_id 
		AND is_deleted = FALSE
`

func (r *postgresRepo) FindAll(userID string) ([]model.UserURLsResponse, error) {
	args := pgx.NamedArgs{"user_id": userID}
	return pg.QueryMany(r.db, findAllURLsSQL, pg.IfaceListFunc[*model.UserURLsResponse](), args)
}

func (r *postgresRepo) Ping() error {
	return r.db.Ping()
}

func (r *postgresRepo) DoTx(f func(r *postgresRepo) error, opts ...*pgx.TxOptions) error {
	return r.db.DoTx(func(mainDb *pg.DB) error {
		return f(NewPostgresRepo(mainDb))
	}, opts...)
}

func (r *postgresRepo) Close() error {
	close(r.deleteCh)
	r.wg.Wait()

	return r.db.Close()
}

const removeUserSQL = `
	UPDATE urls
	SET is_deleted = TRUE
	WHERE short_id = @short_id AND user_id = @user_id AND is_deleted = FALSE
`

func (r *postgresRepo) deletionWorker() {
	defer r.wg.Done()
	buffer := make([]model.DeleteTask, 0, batchSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(buffer) == 0 {
			return
		}

		b := &pgx.Batch{}
		for _, item := range buffer {
			args := pgx.NamedArgs{
				"user_id":  item.UserID,
				"short_id": item.ShortID,
			}
			b.Queue(removeUserSQL, args)
		}

		br, err := r.db.SendBatch(b)
		if err != nil {
			logger.Log.Error("batch delete failed", zap.Error(err))
			return
		}

		for range buffer {
			_, _ = br.Exec()
		}
		br.Close()

		buffer = buffer[:0]
	}

	for {
		select {
		case task, ok := <-r.deleteCh:
			if !ok {
				flush()
				return
			}
			buffer = append(buffer, task)
			if len(buffer) >= batchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func (r *postgresRepo) DeleteURLs(userID string, shortIDs []string) error {
	for _, id := range shortIDs {
		r.deleteCh <- model.DeleteTask{UserID: userID, ShortID: id}
	}

	return nil
}
