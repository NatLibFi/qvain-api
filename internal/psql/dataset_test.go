package psql

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/CSCfi/qvain-api/pkg/models"
	"github.com/wvh/uuid"
)

var owner = uuid.MustFromString("053bffbcc41edad4853bea91fc42ea18")

// TestDatasetNewVersion tests the creation of a new version based on an existing dataset.
func TestDatasetNewVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	type inner struct {
		Title   string `json:"title"`
		Version int    `json:"version"`
	}

	db, err := NewPoolServiceFromEnv()
	if err != nil {
		t.Fatal("psql:", err)
	}

	var v1 uuid.UUID
	t.Run("create", func(t *testing.T) {
		blob := []byte(`{"title":"test dataset","version":1}`)

		dataset, err := models.NewDataset(owner)
		if err != nil {
			t.Fatal("models.NewDataset():", err)
		}
		dataset.SetData(1, "open test dataset", blob)

		v1 = dataset.Id

		err = db.Create(dataset)
		if err != nil {
			t.Fatal("db.Create():", err)
		}
		//defer db.Delete(id, nil)
	})

	var v2 uuid.UUID
	t.Run("version", func(t *testing.T) {
		id, err := uuid.NewUUID()
		if err != nil {
			t.Error(err)
		}

		blob := []byte(`{"title":"test dataset","version":2}`)

		err = db.StoreNewVersion(v1, id, time.Now(), blob)
		if err != nil {
			t.Error(err)
		}
		v2 = id
	})

	t.Run("get", func(t *testing.T) {
		dataset, err := db.Get(v2)
		if err != nil {
			t.Error(err)
		}

		if dataset.Owner.String() != owner.String() {
			t.Errorf("expected owner to be equal; expected %s, got %s", owner, dataset.Owner)
		}

		if !dataset.IsValid() {
			t.Errorf("dataset not marked as valid; expected %v, got %v", true, dataset.IsValid())
		}

		var fields inner
		if err := json.Unmarshal(dataset.Blob(), &fields); err != nil {
			t.Error("error during unmarshal:", err)
		}

		if fields.Version != 2 {
			t.Errorf("field `version` isn't equal; expected %d, got %d", 2, fields.Version)
		}
	})

}
