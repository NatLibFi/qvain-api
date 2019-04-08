package psql

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/CSCfi/qvain-api/pkg/metax"
	"github.com/CSCfi/qvain-api/pkg/models"

	"github.com/wvh/uuid"
)

func readFile(tb testing.TB, fn string) []byte {
	path := filepath.Join("..", "..", "pkg", "metax", "testdata", fn)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		tb.Fatal(err)
	}
	return bytes
}

func createDatasetFromFile(tb testing.TB, db *DB, fn string, owner uuid.UUID) uuid.UUID {
	blob := readFile(tb, fn)

	dataset, err := models.NewDataset(owner)
	if err != nil {
		tb.Fatal("models.NewDataset():", err)
	}
	dataset.SetData(1, metax.SchemaIda, blob)

	err = db.Create(dataset)
	if err != nil {
		tb.Fatal("db.Create():", err)
	}

	return dataset.Id
}

func stringOrNil(val *string) string {
	if val != nil {
		return *val
	} else {
		return "<nil>"
	}
}

type datasetItem struct {
	Identifier *string `json:"identifier"`
	Previous   *string `json:"previous"`
	Next       *string `json:"next"`
}

// TestPublish creates a Qvain dataset, saves it, publishes it to metax, and saves the resulting version.
func TestViewDatasetsByOwner(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tests := []struct {
		published bool
		fn        string
	}{
		{
			published: true,
			fn:        "new_created.json",
		},
		{
			published: false,
			fn:        "unpublished.json",
		},
	}

	owner, err := uuid.NewUUID()
	if err != nil {
		t.Fatal("uuid:", err)
	}

	db, err := NewPoolServiceFromEnv()
	if err != nil {
		t.Fatal("psql:", err)
	}

	// on exit, clean-up datasets by test user
	defer func() {
		if err := db.WithTransaction(func(tx *Tx) error {
			tag, err := tx.Exec("DELETE FROM datasets WHERE owner = $1 AND family = 1", owner.Array())
			if err == nil {
				t.Logf("cleaned up %d datasets", tag.RowsAffected())
			}
			return err
		}); err != nil {
			t.Error("clean-up:", err)
		}
	}()

	for _, test := range tests {
		t.Run(test.fn, func(t *testing.T) {
			id := createDatasetFromFile(t, db, test.fn, owner)
			defer db.Delete(id, nil)

			response, err := db.ViewDatasetsByOwner(owner)
			if err != nil {
				t.Error("view:", response)
			}

			//t.Logf("response: %s\n", response)

			var datasetList []datasetItem
			err = json.Unmarshal(response, &datasetList)
			if err != nil {
				t.Error("json:", err)
			}

			if test.published && datasetList[0].Identifier == nil {
				t.Error("published dataset can't have null identifier")
			} else if !test.published && datasetList[0].Identifier != nil {
				t.Error("unpublished dataset can't have identifier")
			}

			if test.published && datasetList[0].Previous == nil {
				t.Error("new dataset does not have link to previous dataset")
			} else if !test.published && datasetList[0].Previous != nil {
				t.Error("unpublished dataset can't have link to previous dataset")
			} else {
				t.Log("previous dataset:", stringOrNil(datasetList[0].Previous))
			}
		})
	}
}
