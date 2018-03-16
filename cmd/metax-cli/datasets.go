package main

import (
	"fmt"
	"flag"
	"time"
	
	"wvh/uuid/flag"
	"wvh/att/qvain/metax"
)


func runDatasets(url string, args []string) error {
	flags := flag.NewFlagSet("datasets", flag.ExitOnError)
	var (
		owner   uuidflag.Uuid
		since string
		ago   time.Duration
	)
	flags.Var(&owner, "owner", "owner `uuid`")
	flags.StringVar(&since, "since", "", "date in iso-8601 format for Last Modified Since header")
	flags.DurationVar(&ago, "ago", 0, "duration relative to Now() for Last-Modified-Since header, e.g.: \"2h30m\"")
	
	flags.Usage = usageFor(flags, "datasets [flags]")
	if err := flags.Parse(args); err != nil {
		return err
	}
	
	if since != "" && ago != 0 {
		return fmt.Errorf("arguments `since` and `ago` are mutually exclusive")
	}
	
	var sinceHeader time.Time
	
	if since != "" {
		sinceTime, err := time.Parse(time.RFC3339, since)
		if err != nil {
			return fmt.Errorf("can't parse `since` argument: %s", err)
		}
		sinceHeader = sinceTime
	}
	
	if ago != 0 {
		sinceHeader = time.Now().Add(-ago)
	}
	
	if !sinceHeader.IsZero() {
		fmt.Println("Last-Modified-Since:", sinceHeader)
	}
	
	if owner.IsSet() {
		fmt.Println("User:", owner)
	}
	
	fmt.Println("querying metax datasets endpoint")
	svc := metax.NewMetaxService(METAX_HOST)
	// 053bffbcc41edad4853bea91fc42ea18
	response, err := svc.Datasets(metax.WithOwner(owner.String()))
	if err != nil {
		return err
	}
	fmt.Println("count (api):", response.Count)
	for _, rec := range response.Results {
		//fmt.Printf("%+v\n", rec)
		fmt.Printf("%+v\n", rec.Editor)
		if rec.Editor != nil {
			if rec.Editor.OwnerId != nil {
				fmt.Println("owner:", *rec.Editor.OwnerId)
			}
			if rec.Editor.CreatorId != nil {
				fmt.Println("creator:", *rec.Editor.CreatorId)
			}
		}
	}
	
	streamResponse, err := svc.ReadStream(metax.WithOwner(owner.String()))
	if err != nil {
		return err
	}
	
	fmt.Println("stream response:", len(streamResponse))
	for i, dataset := range streamResponse {
 		fmt.Printf("%05d:\n", i+1)
		fmt.Println("  id:", dataset.Id)
		if dataset.Editor != nil {
			fmt.Println("  owner:", *dataset.Editor.OwnerId)
			fmt.Println("  creator:", *dataset.Editor.OwnerId)
			fmt.Println("  identifier:", *dataset.Editor.Identifier)
		}
	}
	
	
	return nil
}
