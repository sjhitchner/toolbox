package db

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// TODO Transactions??

type DBHandler interface {
	DB() *sqlx.DB
	Host() string
	Name() string
	Port() string

	Ping() error
	Close()

	GetById(ctx context.Context, result interface{}, query string, id interface{}) error
	Get(ctx context.Context, results interface{}, query string, params ...interface{}) error
	Select(ctx context.Context, results interface{}, query string, params ...interface{}) error
	InsertWithId(ctx context.Context, query string, params ...interface{}) (int64, error)
	Insert(ctx context.Context, query string, params ...interface{}) error
	Update(ctx context.Context, query string, params ...interface{}) (int64, error)
	Delete(ctx context.Context, query string, params ...interface{}) (int64, error)
}

func WithTransaction(db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {

	tx, err := db.Beginx()
	if err != nil {
		return rollbackError(tx, err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		return rollbackError(tx, err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func rollbackError(tx *sqlx.Tx, err error) error {
	if err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return errors.Wrapf(err, rerr.Error())
		}
		return err
	}
	return nil
}
