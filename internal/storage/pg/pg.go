package pg

import (
	"context"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	pgxdriver "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	stopCtx context.Context
	pool    *pgxpool.Pool
	tx      pgx.Tx
}

func New(ctx context.Context, dsn, migrationsPath string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	if err := applyMigrations(pool, migrationsPath); err != nil {
		pool.Close()
		return nil, err
	}
	return &DB{stopCtx: ctx, pool: pool}, nil
}

func applyMigrations(pool *pgxpool.Pool, migrationsPath string) error {
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	driver, err := pgxdriver.WithInstance(sqlDB, &pgxdriver.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, "postgres", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (db *DB) Close() {
	if db.tx != nil {
		db.tx.Rollback(db.stopCtx)
	}
	db.pool.Close()
}

func (db *DB) WithCtx(ctx context.Context) *DB {
	return &DB{
		stopCtx: ctx,
		pool:    db.pool,
		tx:      db.tx,
	}
}

func (db *DB) DoTx(f func(*DB) error, opts ...*pgx.TxOptions) error {
	if db.tx != nil {
		return f(db)
	}
	var opt *pgx.TxOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	if opt == nil {
		opt = &pgx.TxOptions{
			IsoLevel: pgx.ReadCommitted,
		}
	}
	tx, err := db.pool.BeginTx(db.stopCtx, *opt)
	if err != nil {
		return err
	}
	txDB := &DB{
		stopCtx: db.stopCtx,
		pool:    db.pool,
		tx:      tx,
	}
	if err := f(txDB); err != nil {
		tx.Rollback(db.stopCtx)
		return err
	}
	return tx.Commit(db.stopCtx)
}

func (db *DB) Exec(query string, args ...any) (pgconn.CommandTag, error) {
	if db.tx != nil {
		return db.tx.Exec(db.stopCtx, query, args...)
	}
	return db.pool.Exec(db.stopCtx, query, args...)
}

func (db *DB) Query(query string, args pgx.NamedArgs, f func(pgx.Rows) error) error {
	var (
		rows pgx.Rows
		err  error
	)
	if db.tx != nil {
		rows, err = db.tx.Query(db.stopCtx, query, args)
	} else {
		rows, err = db.pool.Query(db.stopCtx, query, args)
	}
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := f(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (db *DB) QueryRow(query string, args pgx.NamedArgs, dest ...any) error {
	var row pgx.Row
	if db.tx != nil {
		row = db.tx.QueryRow(db.stopCtx, query, args)
	} else {
		row = db.pool.QueryRow(db.stopCtx, query, args)
	}
	return row.Scan(dest...)
}

func (db *DB) Ping() error {
	return db.pool.Ping(db.stopCtx)
}

func (db *DB) SendBatch(batch *pgx.Batch) (pgx.BatchResults, error) {
	if db.tx != nil {
		return db.tx.SendBatch(db.stopCtx, batch), nil
	}
	return db.pool.SendBatch(db.stopCtx, batch), nil
}
