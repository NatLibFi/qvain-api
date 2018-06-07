package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/NatLibFi/qvain-api/metax"
	"github.com/NatLibFi/qvain-api/psql"
	"github.com/wvh/uuid"
	"github.com/wvh/uuid/flag"
)

func runCreate(url string, args []string) error {
	flags := flag.NewFlagSet("create", flag.ExitOnError)
	var (
		owner uuidflag.Uuid
	)
	flags.Var(&owner, "owner", "owner `uuid` to check dataset ownership against")

	flags.Usage = usageFor(flags, "create [flags] <id>")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if flags.NArg() < 1 {
		flags.Usage()
		return fmt.Errorf("error: missing dataset id argument")
	}

	id, err := uuid.FromString(flags.Arg(0))
	if err != nil {
		return err
	}

	if owner.IsSet() {
		fmt.Println("User:", owner)
	}

	psql, err := psql.NewPoolServiceFromEnv()
	if err != nil {
		return err
	}

	dataset, err := psql.Get(id)
	if err != nil {
		return err
	}
	fmt.Println("About to publish:", dataset.Blob())

	api := metax.NewMetaxService(METAX_HOST, metax.WithCredentials(os.Getenv("APP_METAX_API_USER"), os.Getenv("APP_METAX_API_PASS")))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//_, err = api.Create(ctx, []byte(`{}`))
	res, err := api.Create(ctx, dataset.Blob())
	if err != nil {
		fmt.Printf("type: %T\n", err)
		if apiErr, ok := err.(*metax.ApiError); ok {
			fmt.Printf("metax says: %s\n", apiErr.OriginalError())
		}
		return err
	}

	fmt.Printf("metax responded:\n  %s\n", res)

	//id, _ = uuid.FromString("055f200c-52bb-2216-2b54-3752a7466233")
	//owner, _ = uuid.FromString("12345678-9012-3456-7890-123456789012")

	if owner.IsSet() {
		err = psql.MarkPublishedWithOwner(id, owner.Get(), true)
	} else {
		err = psql.MarkPublished(id, true)
	}
	if err != nil {
		return err
	}

	fmt.Println("success")
	return nil
}
