package main

import (
	//"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/CSCfi/qvain-api/internal/psql"
	"github.com/CSCfi/qvain-api/internal/shared"
	"github.com/CSCfi/qvain-api/pkg/metax"
	"github.com/wvh/uuid"
	"github.com/wvh/uuid/flag"
)

func runFetch(url string, args []string) error {
	flags := flag.NewFlagSet("sync", flag.ExitOnError)
	var (
		owner    uuidflag.Uuid
		service  string
		identity string
		since    string
		ago      time.Duration

		uid uuid.UUID
	)
	flags.Var(&owner, "owner", "owner `uuid`")
	flags.StringVar(&service, "service", "fairdata", "external service")
	flags.StringVar(&identity, "identity", "", "external identity")
	flags.StringVar(&since, "since", "", "date in iso-8601 format for Last Modified Since header")
	flags.DurationVar(&ago, "ago", 0, "duration relative to Now() for Last-Modified-Since header, e.g.: \"2h30m\"")

	flags.Usage = usageFor(flags, "fetch [flags]")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if !owner.IsSet() && identity == "" {
		return fmt.Errorf("provide at least one of -owner or -identity")
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
	//fmt.Println(db.Version())

	if identity != "" && !owner.IsSet() {
		uid, err = db.GetUidForIdentity(service, identity)
		if err != nil {
			return err
		}
	} else if identity == "" && owner.IsSet() {
		identity, err = db.GetIdentityForUid(service, owner.Get())
		if err != nil {
			return err
		}
		uid = owner.Get()
	}
	fmt.Printf("request to sync for uid %v identity %q at service %q\n", uid, identity, service)

	api := metax.NewMetaxService(METAX_HOST, metax.WithCredentials(os.Getenv("APP_METAX_API_USER"), os.Getenv("APP_METAX_API_PASS")))

	err = shared.FetchSince(api, db, Logger, uid, identity, sinceHeader)
	if err != nil {
		return err
	}

	fmt.Println("success")
	return nil
}
