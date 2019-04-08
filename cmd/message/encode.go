package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/CSCfi/qvain-api/internal/secmsg"
)

const DefaultKeyEnvName = "APP_TOKEN_KEY"

func getTokenKeyFromEnv() []byte {
	key := os.Getenv(DefaultKeyEnvName)
	if key == "" {
		fmt.Fprintf(os.Stderr, "error: can't find secret key; set `%s`\n", DefaultKeyEnvName)
		os.Exit(1)
	}
	b, err := hex.DecodeString(key)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: can't decode secret key:", err)
		os.Exit(1)
	}
	return b
}

func isPipe(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		return false
	}
	return true
}

func main() {
	msger, err := secmsg.NewMessageService(getTokenKeyFromEnv())
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: can't decode create message service:", err)
		os.Exit(1)
	}

	if !isPipe(os.Stdin) {
		fmt.Fprintln(os.Stderr, "reading from STDIN...")
	}

	msg, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: can't read from STDIN:", err)
		os.Exit(1)
	}
	if len(msg) < 1 {
		fmt.Fprintln(os.Stderr, "error: no data, quiting...")
		os.Exit(1)
	}

	fmt.Printf("%s\n", msger.MustEncode(msg))
}
