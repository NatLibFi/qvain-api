package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/NatLibFi/qvain-api/internal/psql"
	"github.com/NatLibFi/qvain-api/metax"
	"github.com/wvh/uuid/flag"
)

func runFetchOld(url string, args []string) error {
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

	fmt.Println("syncing with metax datasets endpoint")
	pg, err := psql.NewPoolServiceFromEnv()
	if err != nil {
		return err
	}
	//fmt.Println(pg.Check())
	fmt.Println(pg.Version())

	batch, err := pg.NewBatch()
	if err != nil {
		return err
	}
	defer batch.Rollback()

	svc := metax.NewMetaxService(METAX_HOST)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	total, c, errc, err := svc.ReadStreamChannel(ctx, metax.WithOwner(owner.String()))
	if err != nil {
		return err
	}

	fmt.Printf("channel response (%d datasets):", total)
	i := 0
	success := false

Done:
	for {
		select {
		case rawRecord, more := <-c:
			if !more {
				success = true
				break Done
			}
			fmt.Printf("%05d:\n", i+1)
			i++
			dataset, err := rawRecord.ToQvain()
			if err != nil {
				//return err
				fmt.Println("  skipping:", err)
				continue
			}
			fmt.Println("  id:", dataset.Id)
			if err = batch.Store(dataset); err != nil {
				fmt.Println("  Store error:", err)
				continue
			}
		case err := <-errc:
			fmt.Println("api error:", ctx.Err())
			return err
		case <-ctx.Done():
			fmt.Println("api timeout:", ctx.Err())
			return err
		}
	}
	if success {
		batch.Commit()
	}
	fmt.Println("success:", success)
	return nil
}
