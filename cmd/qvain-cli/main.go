// Command qvain-cli is a command-line interface to the Qvain backend.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/CSCfi/qvain-api/internal/psql"
	"github.com/CSCfi/qvain-api/internal/version"
)

const ProgramName = "qvain-cli"

func insertFromFile(fn string) {
	blob, err := ioutil.ReadFile(fn)
	if err != nil {
		panic(err)
	}
	fmt.Printf("File contents: %s", blob)
}

func goUsageFor(flags *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s %s\n", ProgramName, short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		flags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
	}
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

func usage() {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  add         add record")
	fmt.Fprintln(os.Stderr, "  export      export record [json]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  api:")
	fmt.Fprintln(os.Stderr, "  view        view datasets by owner [json]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  db          query db version")
	fmt.Fprintln(os.Stderr, "  version     show version tag if compiled in")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "This program outputs valid JSON on STDOUT for many commands.")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <sub-command> [flags]\n", os.Args[0])
		usage()
		os.Exit(1)
	}

	//var run func(*pgx.Conn, []string) error
	var run func(*psql.DB, []string) error
	switch os.Args[1] {
	case "db":
		run = runPgVersion
	case "add":
		run = runAddRecord
	case "view":
		run = runViewDatasetsByOwner
	case "export":
		run = runExportDataset
	case "version":
		if len(version.CommitTag) > 0 {
			fmt.Fprintln(os.Stderr, "qvain-cli", version.CommitTag)
		} else {
			fmt.Fprintln(os.Stderr, "qvain-cli <unknown>")
		}
		return
	default:
		fmt.Printf("%s: unknown sub-command: %s\n", os.Args[0], os.Args[1])
		usage()
		os.Exit(1)
	}

	pg, err := psql.NewPoolServiceFromEnv()
	if err != nil {
		panic(err)
	}

	if err := run(pg, os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}
