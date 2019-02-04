package main

import (
	"os"

	"github.com/NatLibFi/qvain-api/internal/selfcheck"
)

func main() {
	w := &selfcheck.ConsoleWriter{os.Stdout}
	os.Exit(selfcheck.Run(w))
}
