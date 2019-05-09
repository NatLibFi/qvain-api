package shared

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/CSCfi/qvain-api/internal/psql"
	"github.com/CSCfi/qvain-api/pkg/metax"
	"github.com/wvh/uuid"
)

var (
	// ErrNoIdentifier means we can't find the Metax dataset identifier in created or updated datasets.
	ErrNoIdentifier = errors.New("no identifier in dataset")
)

// Publish stores a dataset in Metax and updates the Qvain database.
// It returns the Metax identifier for the dataset, the new version idenifier if such was created, and an error.
// The error returned can be a Metax ApiError, a Qvain database error, or a basic Go error.
func Publish(api *metax.MetaxService, db *psql.DB, id uuid.UUID, owner uuid.UUID) (versionId string, newVersionId string, newQVersionId *uuid.UUID, err error) {
	/*
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if err := db.CheckOwner(id, owner); err := nil {
			return err
		}
	*/
	dataset, err := db.GetWithOwner(id, owner)
	if err != nil {
		//return err
		return
	}

	fmt.Fprintln(os.Stderr, "About to publish:", id)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := api.Store(ctx, dataset.Blob())
	if err != nil {
		fmt.Fprintf(os.Stderr, "type: %T\n", err)
		if apiErr, ok := err.(*metax.ApiError); ok {
			fmt.Fprintf(os.Stderr, "metax error: [%d] %s\n", apiErr.StatusCode(), apiErr.OriginalError())
		}
		//return err
		return
	}

	fmt.Fprintln(os.Stderr, "Success! Response follows:")
	fmt.Printf("%s", res)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Marking dataset as published...")

	versionId = metax.GetIdentifier(res)
	if versionId == "" {
		return "", "", nil, ErrNoIdentifier
	}

	synced := metax.GetModificationDate(res)
	if synced.IsZero() {
		fmt.Fprintln(os.Stderr, "Could not find date_modified or date_created from dataset!")
		synced = time.Now()
	}

	err = db.StorePublished(id, res, synced)
	if err != nil {
		//return err
		return
	}

	if newVersionId = metax.MaybeNewVersionId(res); newVersionId != "" {
		fmt.Println("created new version:", newVersionId)

		var newVersion []byte
		// get the new version from the Metax api
		newVersion, err = api.GetId(newVersionId)
		if err != nil {
			fmt.Println("error getting new version:", err)
			//return err
			return versionId, newVersionId, nil, err
		}
		fmt.Printf("new: %s\n\n", newVersion)

		// create a Qvain id for the new version
		var tmp uuid.UUID
		tmp, err = uuid.NewUUID()
		if err != nil {
			return
		}
		newQVersionId = &tmp

		synced := metax.GetModificationDate(newVersion)
		if synced.IsZero() {
			fmt.Fprintln(os.Stderr, "Could not find date_modified or date_created from new version!")
			synced = time.Now()
		}

		// store the new version
		err = db.WithTransaction(func(tx *psql.Tx) error {
			return tx.StoreNewVersion(id, *newQVersionId, synced, newVersion)
		})
		if err != nil {
			return
		}
	}

	fmt.Fprintln(os.Stderr, "success")
	return
}
