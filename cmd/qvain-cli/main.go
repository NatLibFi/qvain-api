// Command qvain-cli is a command-line interface to the Qvain backend.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/NatLibFi/qvain-api/psql"
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
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s %s\n", ProgramName, short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		flags.VisitAll(func(f *flag.Flag) {
			//fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
			fType, fUsage := flag.UnquoteUsage(f)
			//defString := fmt.Sprintf(" (default: %v)", f.DefValue)
			//fmt.Fprintf(w, "| type: %s, usage: %s, default: %s (%T)\n", fType, fUsage, defString, f.DefValue)
			if f.DefValue != "" {
				fmt.Fprintf(w, "\t-%s %s\t%s (default: %q)\n", f.Name, fType, fUsage, f.DefValue)
			} else {
				fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, fType, fUsage)
			}
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  add         add record")
	fmt.Fprintln(os.Stderr, "  list        list records")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  version     query Postgresql version")
	fmt.Fprintln(os.Stderr, "")
}

func main() {
	fmt.Println("psql tester")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <sub-command> [flags]\n", os.Args[0])
		usage()
		os.Exit(1)
	}

	//var run func(*pgx.Conn, []string) error
	var run func(*psql.PsqlService, []string) error
	switch os.Args[1] {
	case "version":
		run = runPgVersion
	case "add":
		run = runAddRecord
	default:
		fmt.Printf("%s: unknown sub-command: %s\n", os.Args[0], os.Args[1])
		usage()
		os.Exit(1)
	}

	pg, err := psql.NewService("user=qvain password=" + os.Getenv("PGPASS") + " host=/home/wouter/.s.PGSQL.5432 dbname=qvain sslmode=disable")
	if err != nil {
		panic(err)
	}

	//err = pg.InitPool()
	//if err != nil {
	//	panic(err)
	//}

	/*
		conn, err := pg.NewConn()
		if err != nil {
			//panic(err)
			fmt.Fprintln(os.Stderr, "can't connect to database:", err)
		}
		//defer conn.Close()
	*/

	if err := run(pg, os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}
