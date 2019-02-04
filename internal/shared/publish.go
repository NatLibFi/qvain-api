package shared

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/NatLibFi/qvain-api/internal/psql"
	"github.com/NatLibFi/qvain-api/metax"
	"github.com/wvh/uuid"
)

var (
	// ErrNoIdentifier means we can't find the Metax dataset identifier in created or updated datasets.
	ErrNoIdentifier = errors.New("no identifier in dataset")
)

// Publish stores a dataset in Metax and updates the Qvain database.
// It returns the Metax identifier for the dataset, the new version idenifier if such was created, and an error.
// The error returned can be a Metax ApiError, a Qvain database error, or a basic Go error.
func Publish(api *metax.MetaxService, db *psql.DB, id uuid.UUID, owner uuid.UUID) (versionId string, newVersionId string, err error) {
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
		return "", "", ErrNoIdentifier
	}

	err = db.StorePublished(id, res)
	if err != nil {
		//return err
		return
	}

	/*
		//err = psql.MarkPublishedWithOwner(id, owner.Get(), true)
		err = db.MarkPublished(id, true)
		if err != nil {
			return err
		}
	*/

	if newVersionId = metax.MaybeNewVersionId(res); newVersionId != "" {
		fmt.Println("created new version:", newVersionId)
		newVersion, err := api.GetId(newVersionId)
		if err != nil {
			fmt.Println("error getting new version:", err)
			//return err
			return versionId, newVersionId, err
		}
		fmt.Printf("new: %s\n\n", newVersion)
	}

	fmt.Fprintln(os.Stderr, "success")
	return
}
