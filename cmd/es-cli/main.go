// Command es-cli is a command-line interface to sync reference metadata stored in ElasticSearch.
//
// Example URL: https://metax-test.csc.fi/es/reference_data/funder_type/_search?size=10000&pretty=1&filter_path=hits.hits._source
//
package main

import (
	"fmt"
	"os"

	"github.com/CSCfi/qvain-api/internal/es"
)

func main() {
	//fmt.Println("elastic search query tool")
	esClient := es.NewClient("https://metax-test.csc.fi/es/")
	res, err := esClient.All("reference_data", "funder_type")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//fmt.Printf("%s\n", res)
	//fmt.Println("-- ")

	fmt.Printf("%s\n", es.Filter(res))
	/*
		blob, count := es.FilterAndCount(res)
		fmt.Println("found:", count)
		fmt.Printf("blob: %s\n", blob)
	*/
}
