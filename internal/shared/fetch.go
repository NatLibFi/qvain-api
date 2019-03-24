package shared

import (
	"context"
	"fmt"
	"time"

	"github.com/NatLibFi/qvain-api/internal/psql"
	"github.com/NatLibFi/qvain-api/metax"

	"github.com/rs/zerolog"
	"github.com/wvh/uuid"
)

const DefaultRequestTimeout = 10 * time.Second
const RetryInterval = 10 * time.Second

func Fetch(api *metax.MetaxService, db *psql.DB, logger zerolog.Logger, uid uuid.UUID, extid string) error {
	last, err := db.GetLastSync(uid)
	if err != nil && err != psql.ErrNotFound {
		fmt.Printf("%T %+v\n", err, err)
		return err
	} else if time.Now().Sub(last) < RetryInterval {
		return fmt.Errorf("too soon")
	}

	return fetch(api, db, logger, uid, extid, last)
}

func FetchSince(api *metax.MetaxService, db *psql.DB, logger zerolog.Logger, uid uuid.UUID, extid string, since time.Time) error {
	return fetch(api, db, logger, uid, extid, since)
}

func FetchAll(api *metax.MetaxService, db *psql.DB, logger zerolog.Logger, uid uuid.UUID, extid string) error {
	return fetch(api, db, logger, uid, extid, time.Time{})
}

func fetch(api *metax.MetaxService, db *psql.DB, logger zerolog.Logger, uid uuid.UUID, extid string, since time.Time) error {
	var params []metax.DatasetOption

	// build query options
	if extid == "" {
		// search by Qvain owner
		params = append(params, metax.WithOwner(uid.String()))
	} else {
		// search by external user identity
		params = append(params, metax.WithUser(extid))
	}

	if !since.IsZero() {
		params = append(params, metax.Since(since))
	}

	fmt.Println("syncing with metax datasets endpoint")
	// setup DB batch transaction
	batch, err := db.NewBatchForUser(uid)
	if err != nil {
		return err
	}
	defer batch.Rollback()

	ctx, cancel := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	defer cancel()

	// make API request
	total, c, errc, err := api.ReadStreamChannel(ctx, params...)
	if err != nil {
		return err
	}

	fmt.Printf("channel response will contain %d datasets\n", total)
	logger.Debug().Str("user", uid.String()).Str("identity", extid).Int("count", total).Msg("starting sync with metax")
	read := 0
	written := 0
	success := false

	// loop until all read, error or timeout
Done:
	for {
		select {
		case fdDataset, more := <-c:
			if !more {
				success = true
				break Done
			}
			read++
			fmt.Printf("%05d:\n", read)
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
			written++
		case err := <-errc:
			// error while streaming
			fmt.Println("api error:", ctx.Err())
			return err
		case <-ctx.Done():
			// timeout
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
	fmt.Println("success:", success, read, written)
	return nil
}
