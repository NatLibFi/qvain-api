package main

import (
	"flag"
	"fmt"

	//"github.com/CSCfi/qvain-api/models"
	"github.com/CSCfi/qvain-api/internal/psql"
	"github.com/wvh/uuid"
)

func runExportDataset(psql *psql.DB, args []string) error {
	flags := flag.NewFlagSet("export", flag.ExitOnError)

	flags.Usage = usageFor(flags, "export <id>")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if flags.NArg() < 1 {
		flags.Usage()
		return fmt.Errorf("error: missing <id> parameter")
	}

	id, err := uuid.FromString(flags.Arg(0))
	if err != nil {
		flags.Usage()
		return fmt.Errorf("error: invalid id: %s", err)
	}

	blob, err := psql.ExportAsJson(id)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", blob)

	return nil
}
