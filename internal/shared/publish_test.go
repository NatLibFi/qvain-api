package shared

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/CSCfi/qvain-api/internal/psql"
	"github.com/CSCfi/qvain-api/pkg/metax"
	"github.com/CSCfi/qvain-api/pkg/models"

	"github.com/wvh/uuid"
)

var owner = uuid.MustFromString("053bffbcc41edad4853bea91fc42ea18")

func readFile(tb testing.TB, fn string) []byte {
	path := filepath.Join("..", "..", "pkg", "metax", "testdata", fn)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		tb.Fatal(err)
	}
	return bytes
}

func modifyTitleFromDataset(db *psql.DB, id uuid.UUID, title string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ct, err := tx.Exec("UPDATE datasets SET blob = jsonb_set(blob, '{research_dataset,title,en}', to_jsonb($2::text)), modified = now() WHERE id = $1", id.Array(), title)
	if err != nil {
		return err
	}

	if ct.RowsAffected() != 1 {
		return psql.ErrNotFound
	}

	return tx.Commit()
}

func deleteFilesFromFairdataDataset(db *psql.DB, id uuid.UUID) error {
	return deletePathFromDataset(db, id, "{research_dataset,files}")
}

func deletePathFromDataset(db *psql.DB, id uuid.UUID, path string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ct, err := tx.Exec("UPDATE datasets SET blob = blob #- $2, modified = now() WHERE id = $1", id.Array(), path)
	if err != nil {
		return err
	}

	if ct.RowsAffected() != 1 {
		return psql.ErrNotFound
	}

	return tx.Commit()
}

// TestPublish creates a Qvain dataset, saves it, publishes it to metax, and saves the resulting version.
func TestPublish(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tests := []struct {
		fn string
	}{
		{
			fn: "unpublished.json",
		},
	}

	db, err := psql.NewPoolServiceFromEnv()
	if err != nil {
		t.Fatal("psql:", err)
	}

	api := metax.NewMetaxService(os.Getenv("APP_METAX_API_HOST"), metax.WithCredentials(os.Getenv("APP_METAX_API_USER"), os.Getenv("APP_METAX_API_PASS")))

	for _, test := range tests {
		blob := readFile(t, test.fn)

		dataset, err := models.NewDataset(owner)
		if err != nil {
			t.Fatal("models.NewDataset():", err)
		}
		dataset.SetData(2, metax.SchemaIda, blob)

		id := dataset.Id

		err = db.Create(dataset)
		if err != nil {
			t.Fatal("db.Create():", err)
		}
		//defer db.Delete(id, nil)

		var versionId string

		t.Run(test.fn+"(new)", func(t *testing.T) {
			vId, nId, _, err := Publish(api, db, id, owner)
			if err != nil {
				if apiErr, ok := err.(*metax.ApiError); ok {
					t.Errorf("API error: [%d] %s", apiErr.StatusCode(), apiErr.Error())
				}
				t.Error("error:", err)
			}

			if nId != "" {
				t.Errorf("API created a new version: expected %q, got %q", "", nId)
			}

			t.Logf("published with version id %q", vId)
			versionId = vId
		})

		err = modifyTitleFromDataset(db, id, "Less Wonderful Title")
		if err != nil {
			t.Fatal("modifyTitleFromDataset():", err)
		}

		t.Run(test.fn+"(update)", func(t *testing.T) {
			vId, nId, _, err := Publish(api, db, id, owner)
			if err != nil {
				if apiErr, ok := err.(*metax.ApiError); ok {
					t.Errorf("API error: [%d] %s", apiErr.StatusCode(), apiErr.Error())
				}
				t.Error("error:", err)
			}

			if nId != "" {
				t.Errorf("API created a new version: expected %q, got %q", "", nId)
			}

			if vId != versionId {
				t.Errorf("API version id changed: expected %q, got %q", versionId, vId)
			}

			t.Logf("(re)published with version id %q", vId)
		})

		err = deleteFilesFromFairdataDataset(db, id)
		if err != nil {
			t.Fatal("deleteFilesFromFairdataDataset():", err)
		}

		t.Run(test.fn+"(files)", func(t *testing.T) {
			vId, nId, qId, err := Publish(api, db, id, owner)
			if err != nil {
				if apiErr, ok := err.(*metax.ApiError); ok {
					t.Errorf("API error: [%d] %s", apiErr.StatusCode(), apiErr.Error())
				}
				t.Error("error:", err)
			}

			if vId != versionId {
				t.Errorf("API version id changed: expected %q, got %q", versionId, vId)
			}

			if nId == "" {
				t.Errorf("API didn't create a new version: expected identifier, got %q", nId)
			} else {
				t.Logf("created new version with metax id %q and qvain id %q", nId, qId)
			}

			t.Logf("(re)published with version id %q", vId)
		})

		// if we want to make it invalid again...
		/*
			err = deletePathFromDataset(db, id, "{research_dataset,title,en}")
			if err != nil {
				t.Fatal("deletePathFromDataset():", err)
			}
		*/

	}
}
