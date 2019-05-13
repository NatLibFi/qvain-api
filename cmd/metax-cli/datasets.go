package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"os"

	"github.com/CSCfi/qvain-api/pkg/metax"
	uuidflag "github.com/wvh/uuid/flag"
)

var stringIfMissing = ""

func stringOr(s *string) *string {
	if s != nil {
		return s
	}
	return &stringIfMissing
}

func runDatasets(url string, args []string) error {
	flags := flag.NewFlagSet("datasets", flag.ExitOnError)
	var (
		owner uuidflag.Uuid
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
	svc := metax.NewMetaxService(METAX_HOST, metax.WithCredentials(os.Getenv("APP_METAX_API_USER"), os.Getenv("APP_METAX_API_PASS")))
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
			fmt.Printf("  dump: %#v\n", dataset.Editor)
			fmt.Println("  owner:", *stringOr(dataset.Editor.OwnerId))
			fmt.Println("  creator:", *stringOr(dataset.Editor.OwnerId))
			fmt.Println("  identifier:", *stringOr(dataset.Editor.Identifier))
		}
	}

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
			dataset, err := rawRecord.Record()
			if err != nil {
				//return err
				fmt.Println("  skipping:", err)
				continue
			}
			fmt.Println("  id:", dataset.Id)
			if dataset.Editor != nil {
				//fmt.Printf("  dump: %#v\n", dataset.Editor)
				fmt.Println("  owner:", *stringOr(dataset.Editor.OwnerId))
				fmt.Println("  creator:", *stringOr(dataset.Editor.OwnerId))
				fmt.Println("  identifier:", *stringOr(dataset.Editor.Identifier))
			}
			//i++
		case err := <-errc:
			fmt.Println("api error:", ctx.Err())
			return err
		case <-ctx.Done():
			fmt.Println("api timeout:", ctx.Err())
			return err
		}
	}

	fmt.Println("Done.")
	fmt.Println("success:", success)
	return nil
}
