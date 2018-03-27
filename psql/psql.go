// Package psql is the Postgresql storage layer of Qvain.
package psql

import (
	"github.com/jackc/pgx"
)

type PsqlService struct {
	config *pgx.ConnConfig
	//poolConfig *pgx.ConnPoolConfig
	pool *pgx.ConnPool
}

// NewService returns a database handle with the requested configuration.
// It does not try to connect.
func NewService(connString string) (*PsqlService, error) {
	connConfig, err := pgx.ParseConnectionString(connString)
	if err != nil {
		return nil, err
	}
	return &PsqlService{config: &connConfig}, nil
}

// NewPoolService calls NewService and initialises the connection pool.
func NewPoolService(connString string) (db *PsqlService, err error) {
	db, err = NewService(connString)
	if err != nil {
		return
	}

	err = db.InitPool()
	return
}

func NewPoolServiceFromEnv() (db *PsqlService, err error) {
	connConfig, err := pgx.ParseEnvLibpq()
	if err != nil {
		return nil, err
	}
	db = &PsqlService{config: &connConfig}

	err = db.InitPool()
	return
}

// Connect returns a single database conn or an error.
func (psql *PsqlService) Connect() (*pgx.Conn, error) {
	return pgx.Connect(*psql.config)
}

// MustConnect returns a single database conn and panics on failure.
func (psql *PsqlService) MustConnect() *pgx.Conn {
	conn, err := pgx.Connect(*psql.config)
	if err != nil {
		panic(err)
	}
	return conn
}

// InitPool initialises a pool with default settings on the database object.
func (psql *PsqlService) InitPool() (err error) {
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

func (psql *PsqlService) Begin() (*Tx, error) {
	tx, err := psql.pool.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx}, nil
}

func (psql *PsqlService) Version() (string, error) {
	var version string

	err := psql.pool.QueryRow("select version()").Scan(&version)
	return version, err
}
