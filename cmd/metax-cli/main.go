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
	"fmt"
	"os"
	"flag"
	"text/tabwriter"
	"github.com/NatLibFi/qvain-api/version"
)

const ProgramName = "metax-query"
const API_URL = "https://metax-test.csc.fi/rest/datasets/"
const (
	METAX_HOST = "metax-test.csc.fi"
	
	DATASETS_URL = "/rest/datasets/"
	VERSION_URL = "/rest/version"
)

var Verbose bool

func usage() {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "  datasets   query dataset API endpoint")
	fmt.Fprintln(os.Stderr, "  version    query version")
	fmt.Fprintln(os.Stderr, "")
}


func usageFor(flags *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s %s\n", ProgramName, short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		flags.VisitAll(func(f *flag.Flag) {
			fType, fUsage := flag.UnquoteUsage(f)
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

/*
func notmain() {
	var uid uuidflag.Uuid
	
	flag.BoolVar(&Verbose, "v", false, "verbose")
	flag.Var(&uid, "uid", "user `uuid`")
	flag.Parse()
	
	if len(flag.Args()) != 1 || flag.Arg(0) == "" {
		fmt.Println("MetaX query tester")
		fmt.Println("using:", API_URL)
		fmt.Println()
		fmt.Printf( "usage: %s <recordset id>\n", os.Args[0])
		fmt.Printf( "example: %s pid:urn:cr3\n", os.Args[0])
		os.Exit(1)
	}
	
	api := metax.MetaxServer{ApiUrl: API_URL}
	json, err := api.GetId(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
	if Verbose {
		printBlock("FULL JSON", json)
	}
	
	record, err := api.ParseRecord(json)
	if err != nil {
		panic(err)
	}
	
	if (record.ModifiedByUserId != nil) {
		fmt.Printf("modified_by_user_id: %s\n", *record.ModifiedByUserId)
	}
	if (record.CreatedByApi != nil) {
		fmt.Printf("created_by_api: %s\n", *record.CreatedByApi)
	}
	
	if Verbose {
		printBlock("RESEARCH DATASET", string(record.ResearchDataset))
	}
	
	keys, err := api.ParseRootKeys(json)
	if err != nil {
		panic(err)
	}
	
	fmt.Println("top-level keys:", keys)
	
	_ = time.Now()
	//oneHourAgo := time.Now().Add(-time.Duration(time.Hour))
	//json, err = api.Query(api.UrlForIds(), metax.Since(oneHourAgo), metax.Owner("fucking_test_user"))
	json, err = api.Query(api.UrlForIds(), metax.Owner("fucking_test_user"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//fmt.Println(json)
	list, err := api.ParseList(json)
	if err != nil {
		panic(err)
	}
	fmt.Println("records:", list.Count)
	//fmt.Println("smt:", *list.Results[0].CreatedByApi)
	for i, rec := range list.Results {
		fmt.Printf("result: %d, date: %s\n", i, *rec.CreatedByApi)
	}
}
*/
