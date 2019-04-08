package main

import (
	"os"

	"github.com/CSCfi/qvain-api/internal/selfcheck"
)

func main() {
	w := &selfcheck.ConsoleWriter{os.Stdout}
	os.Exit(selfcheck.Run(w))
}
