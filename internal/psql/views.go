package psql

import (
	"encoding/json"

	"github.com/CSCfi/qvain-api/pkg/models"
	"github.com/wvh/uuid"
)

// apiEmptyList ensures an array is returned even if there are no results.
var apiEmptyList = json.RawMessage([]byte(`[]`))

// ViewDatasetsByOwner builds a JSON array with the datasets for a given owner.
func (db *DB) ViewDatasetsByOwner(owner uuid.UUID) (json.RawMessage, error) {
	var result json.RawMessage

	// note that if there are no results, json_agg will return NULL;
	// we could also catch NULLs by wrapping json_agg with coalesce: coalesce(json_agg(result), '[]')
	err := db.pool.QueryRow(`
		SELECT json_agg(result) "by_owner"
		FROM (
			SELECT id, owner, created, modified, seq, published,
				blob#>'{identifier}' identifier,
				blob#>'{research_dataset,title}' title,
				blob#>'{research_dataset,description}' description,
				blob#>'{preservation_state}' preservation_state,
				blob#>'{previous_dataset_version,identifier}' previous,
				blob#>'{next_dataset_version,identifier}' "next",
				jsonb_array_length(coalesce(blob#>'{dataset_version_set}', '[]')) versions
			FROM datasets
			WHERE owner = $1
		) result
	`, owner.Array()).Scan(&result)
	if err != nil {
		return apiEmptyList, handleError(err)
	}

	// this shouldn't happen, result should be a literal null or array; return an error?
	if len(result) < 1 {
		return apiEmptyList, nil
	}

	// catch null when there are no results (or use coalesce in SQL);
	// comment this out if NULL is preferred over empty list
	if result[0] != '[' {
		return apiEmptyList, nil
	}

	return result, nil
}

// ViewVersions returns a (JSON) array with existing versions for a given dataset and owner.
func (db *DB) ViewVersions(owner uuid.UUID, dataset uuid.UUID) (json.RawMessage, error) {
	var (
		isOwner   bool
		jsonArray json.RawMessage
	)

	err := db.pool.QueryRow(
		`SELECT owner = $1 "is_owner", CASE WHEN owner = $1 AND jsonb_array_length(blob->'dataset_version_set') > 0 THEN blob->'dataset_version_set' ELSE '[]'::jsonb END versions FROM datasets WHERE id = $2`,
		owner.Array(),
		dataset.Array(),
	).Scan(&isOwner, &jsonArray)
	if err != nil {
		return apiEmptyList, handleError(err)
	}

	if !isOwner {
		return apiEmptyList, ErrNotOwner
	}

	return jsonArray, nil
}

func (db *DB) ViewDatasetWithOwner(id uuid.UUID, owner uuid.UUID, svc string) (json.RawMessage, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.CheckOwner(id, owner)
	if err != nil {
		return nil, err
	}

	famId, err := tx.getFamily(id)
	if err != nil {
		return nil, err
	}

	family, err := models.LookupFamily(famId)
	if err != nil {
		return nil, err
	}

	if family.IsPartial() {
		return tx.viewDataset(id, family.Key(), svc)
	}
	return tx.viewDataset(id, "", svc)
}

func (tx *Tx) viewDataset(id uuid.UUID, key string, svc string) (json.RawMessage, error) {
	var (
		record json.RawMessage
		err    error
	)

	// annoyingly similar...
	if key == "" {
		err = tx.QueryRow(`
		SELECT row_to_json(result) "record"
		FROM (
			SELECT id, created, modified, seq, synced, published,
				family AS type, schema, blob AS dataset,
				(SELECT extids->$2 FROM identities WHERE uid = creator) AS creator,
				(SELECT extids->$2 FROM identities WHERE uid = owner) AS owner
			FROM datasets
			WHERE id = $1) result
		`, id.Array(), svc).Scan(&record)
	} else {
		err = tx.QueryRow(`
		SELECT row_to_json(result) "record"
		FROM (
			SELECT id, created, modified, seq, synced, published,
				family AS type, schema, blob#>$2 AS dataset,
				(SELECT extids->$3 FROM identities WHERE uid = creator) AS creator,
				(SELECT extids->$3 FROM identities WHERE uid = owner) AS owner
			FROM datasets
			WHERE id = $1) result
		`, id.Array(), []string{key}, svc).Scan(&record)
	}
	if err != nil {
		return nil, handleError(err)
	}

	return record, nil
}

func (db *DB) ExportAsJson(id uuid.UUID) (json.RawMessage, error) {
	var dataset json.RawMessage

	err := db.pool.QueryRow(`SELECT row_to_json(datasets) FROM datasets WHERE id = $1`, id.Array()).Scan(&dataset)
	if err != nil {
		return nil, handleError(err)
	}

	return dataset, nil
}
