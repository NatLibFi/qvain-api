package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/CSCfi/qvain-api/internal/psql"
	"github.com/CSCfi/qvain-api/pkg/metax"
	uuidflag "github.com/wvh/uuid/flag"
)

func runAddRecord(psql *psql.DB, args []string) error {
	flags := flag.NewFlagSet("add", flag.ExitOnError)
	var (
		creator uuidflag.Uuid
		owner   uuidflag.Uuid // = uuidflag.DefaultFromString("053bffbcc41edad4853bea91fc42ea18") // 053bffbcc41edad4853bea91fc42ea18
		org     string
		schema  string
	)
	flags.Var(&creator, "creator", "creator `uuid`")
	flags.Var(&owner, "owner", "owner `uuid`")
	flags.StringVar(&org, "org", "organisaatio", "organization name")
	flags.StringVar(&schema, "schema", "metax", "schema identifier for given metadata record")

	flags.Usage = usageFor(flags, "add [flags] <json file>")
	if err := flags.Parse(args); err != nil {
		return err
	}

	if flags.NArg() < 1 {
		flags.Usage()
		return fmt.Errorf("error: missing some required arguments")
	}

	if !creator.IsSet() {
		return fmt.Errorf("error: flag `creator` must be set")
	}

	if schema == "" {
		return fmt.Errorf("error: flag `schema` must be set")
	}

	blob, err := ioutil.ReadFile(flags.Arg(0))
	if err != nil {
		return fmt.Errorf("error: can't read record: %s", err)
	}

	dataset, err := metax.NewMetaxDataset(creator.Get())
	if err != nil {
		return err
	}

	identity, err := psql.GetIdentityForUid("fairdata", creator.Get())
	if err != nil || identity == "" {
		// No fairdata id available, use a fake one
		identity = "qvain-user-" + creator.String()
	}

	extra := map[string]string{
		"identity": identity,
		"org":      org,
	}
	dataset.CreateData(metax.MetaxDatasetFamily, schema, blob, extra)

	err = psql.Create(dataset.Unwrap())
	if err != nil {
		return err
	}

	return nil
}
