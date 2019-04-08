package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/CSCfi/qvain-api/internal/psql"
	"github.com/CSCfi/qvain-api/internal/shared"
	"github.com/CSCfi/qvain-api/pkg/metax"
	"github.com/wvh/uuid"
	"github.com/wvh/uuid/flag"
)

func runPublish(url string, args []string) error {
	flags := flag.NewFlagSet("publish", flag.ExitOnError)
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

	db, err := psql.NewPoolServiceFromEnv()
	if err != nil {
		return err
	}

	api := metax.NewMetaxService(os.Getenv("APP_METAX_API_HOST"), metax.WithCredentials(os.Getenv("APP_METAX_API_USER"), os.Getenv("APP_METAX_API_PASS")))

	vId, nId, qId, err := shared.Publish(api, db, id, owner.Get())
	if err != nil {
		fmt.Fprintf(os.Stderr, "type: %T\n", err)
		if apiErr, ok := err.(*metax.ApiError); ok {
			fmt.Fprintf(os.Stderr, "metax error: %s\n", apiErr.OriginalError())
		}
		if dbErr, ok := err.(*psql.DatabaseError); ok {
			fmt.Fprintf(os.Stderr, "database error: %s\n", dbErr.Error())
		}
		return err
	}

	fmt.Fprintln(os.Stderr, "success")
	fmt.Fprintln(os.Stderr, "metax identifier:", vId)
	if nId != "" {
		fmt.Fprintln(os.Stderr, "metax identifier (new version):", nId)
		fmt.Fprintln(os.Stderr, "qvain identifier (new version):", qId)
	}
	return nil
}
