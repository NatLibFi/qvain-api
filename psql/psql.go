// Package psql is the Postgresql storage layer of Qvain.
package psql

import (
	"context"
	"errors"
	"github.com/jackc/pgx"
	"github.com/rs/zerolog"
	"time"
)

type DB struct {
	config *pgx.ConnConfig
	//poolConfig *pgx.ConnPoolConfig
	pool   *pgx.ConnPool
	logger zerolog.Logger
}

// NewService returns a database handle with the requested configuration.
// It does not try to connect.
func NewService(connString string) (*DB, error) {
	connConfig, err := pgx.ParseConnectionString(connString)
	if err != nil {
		return nil, err
	}
	return &DB{config: &connConfig}, nil
}

// NewPoolService calls NewService and initialises the connection pool.
func NewPoolService(connString string) (db *DB, err error) {
	db, err = NewService(connString)
	if err != nil {
		return
	}

	err = db.InitPool()
	return
}

func NewPoolServiceFromEnv() (db *DB, err error) {
	connConfig, err := pgx.ParseEnvLibpq()
	if err != nil {
		return nil, err
	}
	db = &DB{config: &connConfig}

	err = db.InitPool()
	return
}

// SetLogger assigns a zerolog logger to the database service.
func (psql *DB) SetLogger(logger zerolog.Logger) {
	psql.logger = logger
}

// Connect returns a single database conn or an error.
func (psql *DB) Connect() (*pgx.Conn, error) {
	return pgx.Connect(*psql.config)
}

// MustConnect returns a single database conn and panics on failure.
func (psql *DB) MustConnect() *pgx.Conn {
	conn, err := pgx.Connect(*psql.config)
	if err != nil {
		panic(err)
	}
	return conn
}

// InitPool initialises a pool with default settings on the database object.
func (psql *DB) InitPool() (err error) {
	psql.pool, err = pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: *psql.config,
		//MaxConnections int,            // 5: default
		//AfterConnect func(*Conn) error // function to call on connect
		//AcquireTimeout time.Duration   // busy timeout
	})
	return err
}

type Tx struct {
	*pgx.Tx
}

func (psql *DB) Begin() (*Tx, error) {
	tx, err := psql.pool.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx}, nil
}

func (psql *DB) Version() (string, error) {
	var version string

	err := psql.pool.QueryRow("select version()").Scan(&version)
	return version, err
}

func (psql *DB) Check() error {
	conn, err := psql.pool.Acquire()
	if err != nil {
		return err
	}
	defer psql.pool.Release(conn)
	if !conn.IsAlive() {
		return errors.New("connection is dead")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return conn.Ping(ctx)
}
