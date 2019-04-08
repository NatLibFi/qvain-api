package main

import (
	"flag"
	"fmt"
	"os"

	//"github.com/CSCfi/qvain-api/models"
	"github.com/CSCfi/qvain-api/internal/psql"
	"github.com/wvh/uuid/flag"
)

func runViewDatasetsByOwner(psql *psql.DB, args []string) error {
	flags := flag.NewFlagSet("view_byowner", flag.ExitOnError)
	var (
		owner uuidflag.Uuid // = uuidflag.DefaultFromString("053bffbcc41edad4853bea91fc42ea18") // 053bffbcc41edad4853bea91fc42ea18
		extid string
	)
	flags.Var(&owner, "owner", "owner `uuid`")
	flags.StringVar(&extid, "extid", "", "external id")

	flags.Usage = usageFor(flags, "view [flags]")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if !owner.IsSet() && len(extid) < 1 {
		return fmt.Errorf("error: either flag `owner` or flag `extid` must be set")
	}

	blob, err := psql.ViewDatasetsByOwner(owner.Get())
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "datasets for owner %s:\n", owner.Get())
	fmt.Printf("%s\n", blob)

	return nil
}
