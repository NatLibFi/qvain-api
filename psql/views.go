package psql

import (
	"encoding/json"

	"github.com/wvh/uuid"
)

// apiEmptyList ensures an array is returned even if there are no results.
var apiEmptyList = json.RawMessage([]byte(`[]`))

// ViewByOwner builds a JSON array with the datasets for a given owner.
func (db *DB) ViewByOwner(owner uuid.UUID) (json.RawMessage, error) {
	var result json.RawMessage

	// note that if there are no results, json_agg will return NULL;
	// we could also catch NULLs by wrapping json_agg with coalesce: coalesce(json_agg(result), '[]')
	err := db.pool.QueryRow(`
		SELECT json_agg(result) "by_owner"
		FROM (
			SELECT id, owner, created, modified, blob#>'{research_dataset,title}' title, blob#>'{research_dataset,description}' description, blob#>'{preservation_state}' preservation_state
			FROM datasets
			WHERE owner = $1
		) result
	`, owner.Array()).Scan(&result)
	if err != nil {
		return apiEmptyList, err
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

func (db *DB) ExportAsJson(id uuid.UUID) (json.RawMessage, error) {
	var dataset json.RawMessage

	err := db.pool.QueryRow(`SELECT row_to_json(datasets) FROM datasets WHERE id = $1`, id.Array()).Scan(&dataset)
	if err != nil {
		return nil, handleError(err)
	}

	return dataset, nil
}
