package psql

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	SSLModeDisable SSLMode = "disable"     // No SSL
	SSLModeRequire SSLMode = "require"     // Always SSL (skip verification)
	SSLModeFull    SSLMode = "verify-full" // Always SSL (require verification)
)

type SSLMode string

type PSQLHandler struct {
	conn *sqlx.DB
	host string
	name string
	port int
}

func NewPSQLDBHandlerFromString(str string) (*PSQLHandler, error) {
	u, err := url.Parse(str)
	if err != nil {
		return nil, err
	}

	port := 5432
	if u.Port() != "" {
		p, err := strconv.Atoi(u.Port())
		if err != nil {
			return nil, err
		}
		port = p
	}

	password, _ := u.User.Password()
	sslMode := SSLMode(u.Query().Get("sslmode"))
	switch sslMode {
	case SSLModeRequire:
	case SSLModeDisable:
	case SSLModeFull:
	default:
		sslMode = SSLModeDisable
	}

	return NewPSQLDBHandler(
		u.Hostname(),
		u.Path[1:],
		u.User.Username(),
		password,
		port,
		sslMode,
	)
}

// Postgres Connection Object.  Wraps a sqlx.DB and provides a DBConnection() method to access the
// DB Connection
func NewPSQLDBHandler(host, dbname, user, password string, port int, sslmode SSLMode) (*PSQLHandler, error) {
	dsn := fmt.Sprintf("host=%s user=%s password='%s' dbname=%s port=%d sslmode=%s",
		host,
		user,
		password,
		dbname,
		port,
		sslmode,
	)
	conn, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(20)

	psql := &PSQLHandler{
		conn: conn,
		host: host,
		name: dbname,
		port: port,
	}

	err = psql.Ping()
	if err != nil {
		return nil, err
	}

	return psql, nil
}

func (t *PSQLHandler) DB() *sqlx.DB {
	return t.conn
}

func (t *PSQLHandler) Host() string {
	return t.host
}

func (t *PSQLHandler) Name() string {
	return t.name
}

func (t *PSQLHandler) Port() string {
	return fmt.Sprintf("%d", t.port)
}

func (t *PSQLHandler) Ping() error {
	return t.conn.Ping()
}

func (t *PSQLHandler) Close() {
	if err := t.conn.Close(); err != nil {
		panic(err)
	}
}

func IsDuplicateKey(err error) bool {
	return strings.Contains(err.Error(), "duplicate key")
}

func (t *PSQLHandler) GetById(ctx context.Context, result interface{}, query string, id interface{}) error {

	tx, err := t.conn.Beginx()
	if err != nil {
		return err
	}

	if err := tx.Get(result, query, id); err != nil {
		if sql.ErrNoRows == err {
			return Commit(ctx, tx, nil)
		}
		return Rollback(ctx, tx, err)
	}

	return Commit(ctx, tx, err)
}

func (t *PSQLHandler) Get(ctx context.Context, result interface{}, query string, params ...interface{}) error {

	tx, err := t.conn.Beginx()
	if err != nil {
		return err
	}

	if err := tx.Get(result, query, params...); err != nil {
		if sql.ErrNoRows == err {
			return Commit(ctx, tx, nil)
		}
		return Rollback(ctx, tx, err)
	}

	return Commit(ctx, tx, err)
}

func (t *PSQLHandler) Select(ctx context.Context, results interface{}, query string, params ...interface{}) error {

	tx, err := t.conn.Beginx()
	if err != nil {
		return err
	}

	if err := tx.Select(results, query, params...); err != nil {
		if sql.ErrNoRows == err {
			return Commit(ctx, tx, nil)
		}
		return Rollback(ctx, tx, err)
	}

	return Commit(ctx, tx, err)
}

func (t *PSQLHandler) InsertWithId(ctx context.Context, query string, params ...interface{}) (int64, error) {

	if !strings.Contains(strings.ToUpper(query), "RETURNING") {
		panic(fmt.Sprintf("Query (%s) needs to contain a 'RETURNING id' expression", query))
	}

	tx, err := t.conn.Beginx()
	if err != nil {
		return 0, Rollback(ctx, tx, err)
	}

	var id int64
	if err := tx.QueryRow(query, params...).Scan(&id); err != nil {
		//sql.ErrNoRows == err {
		return 0, Rollback(ctx, tx, err)
	}

	return id, Commit(ctx, tx, err)
}

func (t *PSQLHandler) Insert(ctx context.Context, query string, params ...interface{}) error {

	tx, err := t.conn.Beginx()
	if err != nil {
		return Rollback(ctx, tx, err)
	}

	// TODO result
	if _, err := tx.Exec(query, params...); err != nil {
		//sql.ErrNoRows == err {
		return Rollback(ctx, tx, err)
	}

	return Commit(ctx, tx, err)
}

func (t *PSQLHandler) Update(ctx context.Context, query string, params ...interface{}) (int64, error) {

	tx, err := t.conn.Beginx()
	if err != nil {
		return 0, Rollback(ctx, tx, err)
	}

	result, err := tx.Exec(query, params...)
	if err != nil {
		return 0, Rollback(ctx, tx, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, Rollback(ctx, tx, err)
	}

	return affected, Commit(ctx, tx, err)
}

func (t *PSQLHandler) Delete(ctx context.Context, query string, params ...interface{}) (int64, error) {
	return t.Update(ctx, query, params...)
}

func Commit(ctx context.Context, tx *sqlx.Tx, err error) error {
	if err != nil {
		return Rollback(ctx, tx, err)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "Commit Failed")
	}
	return nil
}

func Rollback(ctx context.Context, tx *sqlx.Tx, err error) error {
	if err := tx.Rollback(); err != nil {
		return errors.Wrap(err, "Rollback Failed")
	}
	return err
}
