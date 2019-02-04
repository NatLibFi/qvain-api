package shared

import (
	"context"
	"fmt"
	"time"

	"github.com/NatLibFi/qvain-api/internal/psql"
	"github.com/NatLibFi/qvain-api/metax"
	"github.com/wvh/uuid"
)

const DefaultRequestTimeout = 10 * time.Second
const RetryInterval = 10 * time.Second

func Fetch(api *metax.MetaxService, db *psql.DB, owner uuid.UUID) error {
	last, err := db.GetLastSync(owner)
	if err != nil && err != psql.ErrNotFound {
		fmt.Printf("%T %+v\n", err, err)
		return err
	} else if time.Now().Sub(last) < RetryInterval {
		return fmt.Errorf("too soon")
	}

	return fetch(api, db, owner, last)
}

func FetchSince(api *metax.MetaxService, db *psql.DB, owner uuid.UUID, since time.Time) error {
	return fetch(api, db, owner, since)
}

func FetchAll(api *metax.MetaxService, db *psql.DB, owner uuid.UUID) error {
	return fetch(api, db, owner, time.Time{})
}

func fetch(api *metax.MetaxService, db *psql.DB, owner uuid.UUID, since time.Time) error {
	var params = []metax.DatasetOption{metax.WithOwner(owner.String())}

	if !since.IsZero() {
		fmt.Println("Last-Modified-Since:", since)
		params = append(params, metax.Since(since))
	}

	/*
		last, err := db.GetLastSync(owner)
		if err != nil {
			return err
		}
		if time.Now().Sub(last) < RetryInterval {
			return fmt.Errorf("too soon")
		}
	*/

	fmt.Println("syncing with metax datasets endpoint")
	batch, err := db.NewBatchForUser(owner)
	if err != nil {
		return err
	}
	defer batch.Rollback()

	ctx, cancel := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	defer cancel()

	// ReadStreamChannel(ctx, params) (chan *MetaxRawRecord, chan error, error)
	//c, errc, err := api.ReadStreamChannel(ctx, metax.WithOwner(owner.String()), extraOpts...)
	c, errc, err := api.ReadStreamChannel(ctx, params...)
	if err != nil {
		return err
	}

	fmt.Println("channel response:")
	i := 0
	count := 0
	success := false

Done:
	for {
		select {
		case fdDataset, more := <-c:
			if !more {
				success = true
				break Done
			}
			fmt.Printf("%05d:\n", i+1)
			i++
			dataset, err := fdDataset.ToQvain()
			if err != nil {
				//return err
				fmt.Println("  skipping:", err)
				continue
			}
			fmt.Println("  id:", dataset.Id)
			if err = batch.Update(dataset.Id, dataset.Blob()); err != nil {
				fmt.Println("  Store error:", err)
				continue
			}
			count++
		case err := <-errc:
			fmt.Println("api error:", ctx.Err())
			return err
		case <-ctx.Done():
			fmt.Println("api timeout:", ctx.Err())
			return err
		}
	}
	if success {
		err = batch.Commit()
	}
	if err != nil {
		return err
	}
	fmt.Println("success:", success, i, count)
	return nil
}
