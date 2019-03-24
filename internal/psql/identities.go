package psql

import (
	"fmt"

	"github.com/wvh/uuid"
)

// RegisterIdentity takes an external service and identity and gets the corresponding application uid, creating it if necessary.
// Returns either a UUID uid and a boolean indicating if the account has been newly created or an error.
//
// Note that while external ids are not guaranteed to be unique and might refer to multiple application users,
// this function allows only unique registrations as it creates real (login) users mapped to external accounts.
func (db *DB) RegisterIdentity(svc, id string) (uid uuid.UUID, isNew bool, err error) {
	var tx *Tx

	tx, err = db.Begin()
	if err != nil {
		fmt.Println("ERROR:", err)
		return uid, isNew, handleError(err)
	}
	defer tx.Rollback()

	// the database can't create our UIDs, so create one in case we need it
	uid, err = uuid.NewUUID()
	if err != nil {
		//return uid, isNew, err
		return
	}

	err = tx.QueryRow(`SELECT uid, is_new FROM register_identity($1, $2, $3)`, uid.Array(), svc, id).Scan(uid.Array(), &isNew)
	if err != nil {
		return uid, isNew, handleError(err)
	}

	return uid, isNew, tx.Commit()
}

// GetUidForIdentity gets the application uid for a given external service and identity.
//
// Note that identities need not be unique, though those used for login ought to be.
func (db *DB) GetUidForIdentity(svc, id string) (uid uuid.UUID, err error) {
	// typecast is necessary for Postgresql to know the data type of variadic arguments in prepared statements
	err = db.pool.QueryRow(`SELECT uid FROM identities WHERE extids @> jsonb_build_object($1::text, $2::text)`, svc, id).Scan(uid.Array())
	if err != nil {
		return uid, handleError(err)
	}

	return uid, nil
}

// GetIdentityForUid gets the identity for a given uid and service.
func (db *DB) GetIdentityForUid(svc string, uid uuid.UUID) (id string, err error) {
	err = db.pool.QueryRow(`SELECT extids->>$1 FROM identities WHERE uid = $2`, svc, uid.Array()).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}
