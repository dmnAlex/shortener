package pg

import (
	"context"
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
)

type DB struct {
	stopCtx context.Context
	db      *sql.DB
	tx      *sql.Tx
}

func New(ctx context.Context, dsn, migrationsPath string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	if err := applyMigrations(db, migrationsPath); err != nil {
		db.Close()
		return nil, err
	}

	return &DB{stopCtx: ctx, db: db}, nil
}

func applyMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, "shortener", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func (db *DB) WithCtx(ctx context.Context) *DB {
	return &DB{
		stopCtx: ctx,
		db:      db.db,
		tx:      db.tx,
	}
}

func (db *DB) DoTx(f func(*DB) error, opts ...*sql.TxOptions) error {
	if db.tx != nil {
		return f(db)
	}

	var opt *sql.TxOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	if opt == nil {
		opt = &sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
		}
	}

	tx, err := db.db.BeginTx(db.stopCtx, opt)
	if err != nil {
		return err
	}

	txDB := &DB{
		stopCtx: db.stopCtx,
		db:      db.db,
		tx:      tx,
	}

	if err := f(txDB); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	if db.tx != nil {
		return db.tx.ExecContext(db.stopCtx, query, args...)
	}
	return db.db.ExecContext(db.stopCtx, query, args...)
}

func (db *DB) Query(query string, args pgx.NamedArgs, f func(*sql.Rows) error) error {
	var (
		rows *sql.Rows
		err  error
	)
	if db.tx != nil {
		rows, err = db.tx.QueryContext(db.stopCtx, query, args)
	} else {
		rows, err = db.db.QueryContext(db.stopCtx, query, args)
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
	if db.tx != nil {
		return db.tx.QueryRowContext(db.stopCtx, query, args).Scan(dest...)
	}
	return db.db.QueryRowContext(db.stopCtx, query, args).Scan(dest...)
}

func (db *DB) Ping() error {
	return db.db.PingContext(db.stopCtx)
}
