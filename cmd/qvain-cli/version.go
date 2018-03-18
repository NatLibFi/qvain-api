
package main

import (
	"fmt"
	
	"github.com/NatLibFi/qvain-api/psql"
)


func runPgVersion(psql *psql.PsqlService, args []string) error {
	var version string
	conn, err := psql.NewConn()
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
