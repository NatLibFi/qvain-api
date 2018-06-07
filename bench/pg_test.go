package main

import (
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx"
)

// DSN="user=username password=password host=1.2.3.4 port=5432 dbname=mydb sslmode=disable"
// "user=qvain password=" + os.Getenv("PGPASS") + " host=/home/wouter/.s.PGSQL.5432 dbname=qvain sslmode=disable"

func BenchmarkSelect(b *testing.B) {
	var query string = "SELECT schema FROM datasets WHERE id = '055f1f96-1d1d-e046-3457-b15e1bd8c10c'"

	config, err := pgx.ParseEnvLibpq()
	if err != nil {
		b.Fatal("can't parse psql config from env: ", err)
	}

	conn, err := pgx.Connect(config)
	if err != nil {
		b.Fatal("can't connect to postgresql: ", err)
	}

	var tmp string
	start := time.Now()
	err = conn.QueryRow(query).Scan(&tmp)
	if err != nil {
		b.Fatal("query fails: ", err)
	}
	b.Logf("test query took: %v\n", time.Now().Sub(start))

	b.Run("simplestring", func(b *testing.B) {
		var res string
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := conn.QueryRow(query).Scan(&res)
			if err != nil {
				b.Fatal("error during query:", err)
			}
		}
	})
}
