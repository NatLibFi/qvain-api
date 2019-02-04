package main

import (
	//"context"
	"flag"
	"fmt"
	"time"

	"github.com/NatLibFi/qvain-api/internal/psql"
	"github.com/NatLibFi/qvain-api/internal/shared"
	"github.com/NatLibFi/qvain-api/metax"
	"github.com/wvh/uuid/flag"
)

func runFetch(url string, args []string) error {
	flags := flag.NewFlagSet("sync", flag.ExitOnError)
	var (
		owner uuidflag.Uuid
		since string
		ago   time.Duration
	)
	flags.Var(&owner, "owner", "owner `uuid`")
	flags.StringVar(&since, "since", "", "date in iso-8601 format for Last Modified Since header")
	flags.DurationVar(&ago, "ago", 0, "duration relative to Now() for Last-Modified-Since header, e.g.: \"2h30m\"")

	flags.Usage = usageFor(flags, "fetch [flags]")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if !owner.IsSet() {
		return fmt.Errorf("owner is a required parameter")
		//fmt.Println("User:", owner)
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

	fmt.Println("syncing with metax datasets endpoint")
	db, err := psql.NewPoolServiceFromEnv()
	if err != nil {
		return err
	}
	fmt.Println(db.Version())

	api := metax.NewMetaxService(METAX_HOST)

	err = shared.FetchSince(api, db, owner.Get(), sinceHeader)
	if err != nil {
		return err
	}

	fmt.Println("success")
	return nil
}
