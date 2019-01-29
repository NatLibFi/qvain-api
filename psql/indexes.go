package psql

import (
	"github.com/wvh/uuid"
)

func (db *DB) LookupByQvainId(id uuid.UUID) (bool, error) {
	var exists bool
	err := db.pool.QueryRow(`SELECT true FROM datasets WHERE id = $1 LIMIT 1`, id.Array()).Scan(&exists)
	return exists, handleError(err)
}

// LookupByFairdataIdentifier returns the Qvain id for a given Fairdata identifier.
func (db *DB) LookupByFairdataIdentifier(fdid string) (uuid.UUID, error) {
	var id uuid.UUID
	//err := db.pool.QueryRow(`SELECT id FROM datasets WHERE family = 2 AND blob @> '{"identifier": $1}'`, `"` + fdid + `"`).Scan(&id)
	err := db.pool.QueryRow(`SELECT id FROM datasets WHERE family = 2 AND blob @> jsonb_build_object('identifier', $1::text)`, fdid).Scan(id.Array())
	if err != nil {
		return id, handleError(err)
	}

	return id, nil
}
