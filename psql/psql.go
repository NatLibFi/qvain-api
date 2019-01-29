// Package psql is the Postgresql storage layer of Qvain.
package psql

import (
	"context"
	"errors"
	"github.com/jackc/pgx"
	"github.com/rs/zerolog"
	"time"
)

// DefaultPoolAcquireTimeout is the duration pgx waits for a connection to become available from the pool.
const DefaultPoolAcquireTimeout = 10 * time.Second

// DB holds the database methods and configuration.
type DB struct {
	config *pgx.ConnConfig
	//poolConfig *pgx.ConnPoolConfig
	pool   *pgx.ConnPool
	logger zerolog.Logger
}

// NewService returns a database handle configured with the given connection string.
// It does not try to connect.
func NewService(connString string) (db *DB, err error) {
	connConfig, err := pgx.ParseConnectionString(connString)
	if err != nil {
		return nil, err
	}

	db = newService(&connConfig)
	return
}

// newService is the actual constructor that takes a ConnConfig populated by the calling function in whatever way.
func newService(config *pgx.ConnConfig) (db *DB) {
	db = &DB{
		config: config,
		logger: zerolog.Nop(),
	}
	if true {
		// self-referential, should be ok with the garbage collector...
		db.config.Logger = db
	}
	return
}

// NewPoolService creates a psql service from a connection string and initialises the connection pool.
func NewPoolService(connString string) (db *DB, err error) {
	connConfig, err := pgx.ParseConnectionString(connString)
	if err != nil {
		return nil, err
	}

	db = newService(&connConfig)
	err = db.InitPool()
	return
}

// NewPoolService creates a psql service using environment variables and initialises the connection pool.
func NewPoolServiceFromEnv() (db *DB, err error) {
	connConfig, err := pgx.ParseEnvLibpq()
	if err != nil {
		return nil, err
	}

	db = newService(&connConfig)
	err = db.InitPool()
	return
}

// SetLogger assigns a zerolog logger to the database service.
// It is not safe to call this function after initialisation.
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
	// default MaxConnections: 5
	psql.pool, err = pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     *psql.config,
		AcquireTimeout: DefaultPoolAcquireTimeout,
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

func (psql *DB) Log(plevel pgx.LogLevel, msg string, data map[string]interface{}) {
	var zlevel zerolog.Level

	switch plevel {
	case pgx.LogLevelNone:
		zlevel = zerolog.NoLevel
	case pgx.LogLevelError:
		zlevel = zerolog.ErrorLevel
	case pgx.LogLevelWarn:
		zlevel = zerolog.WarnLevel
	case pgx.LogLevelInfo:
		zlevel = zerolog.InfoLevel
	case pgx.LogLevelDebug:
		zlevel = zerolog.DebugLevel
	default:
		zlevel = zerolog.DebugLevel
	}

	pgxlog := psql.logger.With().Fields(data).Logger()
	pgxlog.WithLevel(zlevel).Msg(msg)
}
