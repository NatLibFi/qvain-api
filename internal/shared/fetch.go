package shared

import (
	"context"
	"fmt"
	"time"

	"github.com/CSCfi/qvain-api/internal/psql"
	"github.com/CSCfi/qvain-api/pkg/metax"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/wvh/uuid"
)

const DefaultRequestTimeout = 15 * time.Second
const RetryInterval = 10 * time.Second

func Fetch(api *metax.MetaxService, db *psql.DB, logger zerolog.Logger, uid uuid.UUID, extid string) error {
	last, err := db.GetLastSync(uid)
	if err != nil && err != psql.ErrNotFound {
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

	//logger.Info().Str("user", uid.String()).Str("identity", extid).Int("count", total).Msg("starting sync with metax")

	// create sub-logger to correlate possibly multiple log entries
	syncLogger := logger.With().Str("sync-id", xid.New().String()).Logger()
	syncLogger.Info().Str("user", uid.String()).Str("identity", extid).Int("total", total).Msg("starting sync")

	read := 0
	written := 0
	success := false

	// get existing Qvain datasets for user
	userDatasets, err := db.GetAllForUid(uid)
	if err != nil {
		syncLogger.Error().Err(err).Msg("failed to get user datasets")
	}

	// Map Metax identifier in Qvain dataset to the dataset id.
	// Used when a dataset from Metax does not have a Qvain id in its editor metadata.
	// Also get per-dataset timestamp of last sync.
	metaxDatasetQvainId := make(map[string]*uuid.UUID)
	qvainDatasetSyncTime := make(map[uuid.UUID]time.Time)
	if total > 0 {
		for _, ds := range userDatasets {
			if ds.Family() != metax.MetaxDatasetFamily {
				continue
			}

			qvainDatasetSyncTime[ds.Id] = ds.Synced
			metaxIdentifier := metax.GetIdentifier(ds.Blob())
			if metaxIdentifier == "" {
				continue
			}
			if _, exists := metaxDatasetQvainId[metaxIdentifier]; exists {
				syncLogger.Warn().Str("identifier", metaxIdentifier).Msg("multiple datasets have the same Metax indentifier")
				continue
			}
			metaxDatasetQvainId[metaxIdentifier] = &ds.Id
		}
	}

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

			// create dataset, use Qvain id from editor metadata if available
			dataset, isNew, err := fdDataset.ToQvain()
			if err != nil {
				syncLogger.Debug().Err(err).Int("read", read).Msg("error parsing dataset, skipping")
				continue
			}

			// was the Metax dataset not from Qvain?
			if isNew {
				// check if we already have a dataset with the same Metax identifier
				identifier := metax.GetIdentifier(fdDataset.RawMessage)
				if identifier != "" {
					newId, found := metaxDatasetQvainId[identifier]
					if found {
						// update the existing dataset blob instead of creating a new dataset
						isNew = false
						dataset.Id = *newId
					}
				}
			}

			if isNew {
				// create new id
				dataset.Id, err = uuid.NewUUID()
				if err != nil {
					return err
				}

				// inject current user for datasets created externally
				dataset.Creator = uid
				dataset.Owner = uid

				// dataset comes from upstream, so consider it published and valid
				dataset.Published = true
				dataset.SetValid(true)

				if err = batch.CreateWithMetadata(dataset); err != nil {
					syncLogger.Debug().Err(err).Int("read", read).Str("id", dataset.Id.String()).Msg("can't store dataset")
					continue
				}
			} else {
				// check if we have already synced the Qvain dataset based on modification dates
				modified := metax.GetModificationDate(dataset.Blob())
				if !modified.IsZero() && !modified.After(qvainDatasetSyncTime[dataset.Id]) {
					syncLogger.Debug().Str("id", dataset.Id.String()).Msg("dataset not modified in Metax after last sync")
					if err = batch.UpdateSynced(dataset.Id); err != nil {
						syncLogger.Debug().Err(err).Int("read", read).Str("id", dataset.Id.String()).Msg("could't update sync timestamp")
					}
					continue
				}

				if err = batch.Update(dataset.Id, dataset.Blob()); err != nil {
					syncLogger.Debug().Err(err).Int("read", read).Str("id", dataset.Id.String()).Msg("can't update dataset")
					continue
				}
			}
			syncLogger.Debug().Bool("new", isNew).Str("id", dataset.Id.String()).Msg("batched dataset")
			written++
		case err := <-errc:
			// error while streaming
			syncLogger.Info().Err(err).Msg("api error")
			return err
		case <-ctx.Done():
			// timeout
			syncLogger.Info().Err(ctx.Err()).Msg("api timeout")
			return err
		}
	}
	if success {
		err = batch.Commit()
	}
	if err != nil {
		return err
	}

	syncLogger.Info().Int("total", total).Int("written", written).Msg("successful sync")
	return nil
}
