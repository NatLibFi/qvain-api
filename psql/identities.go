package psql

import (
	"github.com/wvh/uuid"
)

// CreateIdentity creates an external identity. It does nothing if the external identity exists already.
func (db *DB) CreateIdentity(id string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	provisionalUuid, err := uuid.NewUUID()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO users (uid, extid) VALUES ($2, $1) ON CONFLICT (extid) DO NOTHING`, id, provisionalUuid)
	if err != nil {
		return handleError(err)
	}

	return tx.Commit()
}


// CreateOrGetIdentity creates an external identity if it doesn't exist already and returns the internal UUID id.
func (db *DB) CreateOrGetIdentity(id string) (uuid.UUID, error) {
	var newUuid uuid.UUID

	tx, err := db.Begin()
	if err != nil {
		return newUuid, err
	}
	defer tx.Rollback()

	provisionalUuid, err := uuid.NewUUID()
	if err != nil {
		return newUuid, err
	}

	err = tx.QueryRow(`
		WITH existing AS (
			SELECT uid FROM identities WHERE extid=$1
		), inserted AS (
			INSERT INTO users (uid, extid) VALUES ($2, $1)
			ON CONFLICT (extid) DO NOTHING
			RETURNING uid
		)
		SELECT uid
		FROM existing
		UNION ALL
		SELECT uid
		FROM inserted`,
		id, provisionalUuid).Scan(newUuid)

	if err != nil {
		return newUuid, handleError(err)
	}

	return newUuid, tx.Commit()
}
