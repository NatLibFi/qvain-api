// Command metax-cli is a command-line client to the Metax API.
package main

/*
	links:
		https://metax-test.csc.fi
		https://metax-test.csc.fi/rest/datasets/pid:urn:cr3

	curl:
		curl -s https://metax-test.csc.fi/rest/datasets/?owner=053bffbcc41edad4853bea91fc42ea18 | jq -r '"count: "+(.count|tostring),"keys: "+(keys|join(",")),"length: "+(.results|length|tostring)'
		curl -s https://metax-test.csc.fi/rest/datasets/?owner=053bffbcc41edad4853bea91fc42ea18 | jq -r '.count'
*/

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/CSCfi/qvain-api/internal/version"

	"github.com/rs/zerolog"
)

const ProgramName = "metax-query"
const API_URL = "https://metax-test.csc.fi/rest/datasets/"
const (
	METAX_HOST = "metax-test.csc.fi"

	DATASETS_URL = "/rest/datasets/"
	VERSION_URL  = "/rest/version"
)

var (
	Verbose bool
	Logger  zerolog.Logger
)

func init() {
	zerolog.TimeFieldFormat = "15:04:05.000000"
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05.000000"}).With().Caller().Timestamp().Logger()
}

func usage() {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  datasets   query dataset API endpoint")
	fmt.Fprintln(os.Stderr, "  fetch      fetch from dataset API endpoint")
	fmt.Fprintln(os.Stderr, "  publish    publish dataset to API endpoint")
	fmt.Fprintln(os.Stderr, "  version    query version")
	fmt.Fprintln(os.Stderr, "")
}

func usageFor(flags *flag.FlagSet, short string) func() {
	return func() {
		var hasFlags bool = false // go doesn't let us count how many flags are defined

		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s %s\n", ProgramName, short)
		fmt.Fprintf(os.Stderr, "\n")

		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		flags.VisitAll(func(f *flag.Flag) {
			if !hasFlags {
				fmt.Fprintf(os.Stderr, "FLAGS\n")
				hasFlags = true
			}
			fType, fUsage := flag.UnquoteUsage(f)
			if f.DefValue != "" {
				fmt.Fprintf(w, "\t-%s %s\t%s (default: %q)\n", f.Name, fType, fUsage, f.DefValue)
			} else {
				fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, fType, fUsage)
			}
		})
		w.Flush()
		if hasFlags {
			fmt.Fprintf(os.Stderr, "\n")
		}
	}
}

func printBlock(name, block string) {
	fmt.Println("--8<-- BEGIN", name, "------")
	fmt.Println(block)
	fmt.Println("--8<-- END", name, "------")
}

func main() {
	fmt.Printf("%s (version %s)\n", ProgramName, version.CommitHash)

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <sub-command> [flags]\n", os.Args[0])
		usage()
		os.Exit(1)
	}

	var run func(string, []string) error
	var endpoint string

	switch os.Args[1] {
	case "datasets":
		endpoint = DATASETS_URL
		run = runDatasets
	case "fetch":
		endpoint = DATASETS_URL
		run = runFetch
	case "publish":
		endpoint = DATASETS_URL
		run = runPublish
	case "version":
		run = runVersion
	default:
		fmt.Printf("%s: unknown sub-command: %s\n", os.Args[0], os.Args[1])
		usage()
		os.Exit(1)
	}

	if err := run(endpoint, os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
