package main

import (
	"fmt"

	"github.com/CSCfi/qvain-api/internal/psql"
)

func runPgVersion(psql *psql.DB, args []string) error {
	var version string
	conn, err := psql.Connect()
	if err != nil {
		panic(err)
	}
	err = conn.QueryRow("select version()").Scan(&version)
	if err != nil {
		panic(err)
	}
	fmt.Println("Postgresql version:", version)
	return nil
}
