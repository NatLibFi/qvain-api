package psql

import (
	//"errors"

	"log"
	"time"

	"github.com/CSCfi/qvain-api/pkg/models"
	"github.com/wvh/uuid"
)

// Create creates a new dataset. It is a convenience wrapper for the Create method on transactions.
func (db *DB) Create(dataset *models.Dataset) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.Create(dataset)
	if err != nil {
		return handleError(err)
	}

	return tx.Commit()
}

// BatchStore takes a list of datasets and stores them as new datasets.
func (db *DB) BatchStore(datasets []*models.Dataset) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// do something batch-like
	for _, dataset := range datasets {
		err = tx.Create(dataset)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Create inserts a new dataset into the database using default values for date and boolean fields.
// Use this for new datasets created in this application.
func (tx *Tx) Create(dataset *models.Dataset) error {
	_, err := tx.Exec("INSERT INTO datasets(id, creator, owner, family, schema, blob) VALUES($1, $2, $3, $4, $5, $6)",
		dataset.Id.Array(),
		dataset.Creator.Array(),
		dataset.Owner.Array(),
		dataset.Family(),
		dataset.Schema(),
		dataset.Blob(),
	)
	if err != nil {
		return err
	}

	return nil
}

// createWithMetadata inserts a new dataset into the database, but also populates other fields.
// Use this when the new dataset already has some metadata fields set, such as when it origates from other services.
//
// This method does not set Modified, as that field is reserved for user edits.
func (tx *Tx) createWithMetadata(dataset *models.Dataset) error {
	_, err := tx.Exec("INSERT INTO datasets(id, creator, owner, created, synced, published, valid, family, schema, blob) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		dataset.Id.Array(),
		dataset.Creator.Array(),
		dataset.Owner.Array(),
		dataset.Created,
		dataset.Synced,
		dataset.Published,
		dataset.IsValid(),
		dataset.Family(),
		dataset.Schema(),
		dataset.Blob(),
	)
	if err != nil {
		return err
	}

	return nil
}

// StoreNewVersion inserts a new version of an existing dataset, copying most fields.
func (tx *Tx) StoreNewVersion(basedOn uuid.UUID, id uuid.UUID, created time.Time, blob []byte) error {
	tag, err := tx.Exec(`
	INSERT INTO datasets (id, creator, owner, created, synced, published, valid, family, schema, blob)
		SELECT $2, creator, owner, $3, $3, true, true, family, schema, $4
		FROM datasets
		WHERE id = $1
	`, basedOn.Array(), id.Array(), created, blob)
	if err != nil {
		return err
	}

	// if the SELECT doesn't match any record (no parent), INSERT will return 0 (inserted records)
	if tag.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

// WithTransaction abstracts some of the database logic by wrapping Tx methods.
//
// example:
// 	db.WithTransaction(func(tx psql.Tx) error {
// 		return tx.StoreNewVersion(id, parent, created, blob)
// 	})
func (db *DB) WithTransaction(f func(tx *Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = f(tx)
	if err != nil {
		return handleError(err)
	}

	return tx.Commit()
}

// StoreNewVersion wraps a StoreNewVersion transaction.
func (db *DB) StoreNewVersion(id uuid.UUID, basedOn uuid.UUID, created time.Time, blob []byte) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.StoreNewVersion(id, basedOn, created, blob)
	if err != nil {
		return handleError(err)
	}

	return tx.Commit()
}

func (db *DB) Update(id uuid.UUID, blob []byte) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.update(id, blob)
	if err != nil {
		return handleError(err)
	}

	return tx.Commit()
}

// UpdateWithOwner updates a dataset with ownership checks.
func (db *DB) UpdateWithOwner(id uuid.UUID, blob []byte, owner uuid.UUID) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.CheckOwner(id, owner)
	if err != nil {
		return err
	}

	err = tx.update(id, blob)
	if err != nil {
		return handleError(err)
	}

	return tx.Commit()
}

// internal update, user triggered
func (tx *Tx) update(id uuid.UUID, blob []byte) error {
	ct, err := tx.Exec("UPDATE datasets SET modified = now(), seq = seq + 1, blob = $2 WHERE id = $1", id.Array(), blob)
	if err != nil {
		return err
	}

	if ct.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

// internal update synced, service triggered
func (tx *Tx) updateSyncedByService(id uuid.UUID) error {
	ct, err := tx.Exec("UPDATE datasets SET synced = now(), seq = seq + 1 WHERE id = $1", id.Array())
	if err != nil {
		return err
	}

	if ct.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

// internal update, service triggered
func (tx *Tx) updateByService(id uuid.UUID, blob []byte) error {
	ct, err := tx.Exec("UPDATE datasets SET synced = now(), modified = now(), seq = seq + 1, blob = $2 WHERE id = $1", id.Array(), blob)
	if err != nil {
		return err
	}

	if ct.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (db *DB) Patch(id uuid.UUID, blob []byte) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.patch(id, blob)
	if err != nil {
		return handleError(err)
	}

	return tx.Commit()
}

// PatchWithOwner patches a dataset JSON blob with ownership checks.
func (db *DB) PatchWithOwner(id uuid.UUID, blob []byte, owner uuid.UUID) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.CheckOwner(id, owner)
	if err != nil {
		return err
	}

	err = tx.patch(id, blob)
	if err != nil {
		return handleError(err)
	}

	return tx.Commit()
}

func (tx *Tx) patch(id uuid.UUID, blob []byte) error {
	ct, err := tx.Exec("UPDATE datasets SET modified = now(), seq = seq + 1, blob = blob || $2 WHERE id = $1", id.Array(), blob)
	if err != nil {
		return err
	}

	if ct.RowsAffected() != 1 {
		return ErrNotFound
	}

	return nil
}

func (db *DB) SmartGetWithOwner(id uuid.UUID, owner uuid.UUID) (*models.Dataset, error) {
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
		return tx.get(id, family.Key())
	}
	return tx.get(id, "")
}

func (db *DB) SmartUpdateWithOwner(id uuid.UUID, blob []byte, owner uuid.UUID) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.CheckOwner(id, owner)
	if err != nil {
		return err
	}

	famId, err := tx.getFamily(id)
	if err != nil {
		return err
	}

	family, err := models.LookupFamily(famId)
	if err != nil {
		return err
	}

	if family.IsPartial() {
		err = tx.patch(id, blob)
	} else {
		err = tx.update(id, blob)
	}
	if err != nil {
		return handleError(err)
	}

	return tx.Commit()
}

// StorePublished saves a published dataset to the database and marks it as published.
// TODO: handle empty blob
func (db *DB) StorePublished(id uuid.UUID, blob []byte, synced time.Time) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ct, err := tx.Exec("UPDATE datasets SET blob = $2, published = true, synced = $3, seq = seq + 1 WHERE id = $1",
		id.Array(), blob, synced)
	if err != nil {
		return handleError(err)
	}

	if ct.RowsAffected() != 1 {
		return ErrNotFound
	}

	return tx.Commit()
}

func (db *DB) Clone(id uuid.UUID, newid uuid.UUID, blob []byte) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	ct, err := tx.Exec(`
		INSERT INTO datasets(id, creator, owner, created, modified, synced, published, valid, family, schema, blob)
		(SELECT $2, creator, owner, created, modified, synced, published, valid, family, schema, $3 WHERE id = $1)`,
		id, newid, blob)
	if err != nil {
		return handleError(err)
	}

	if ct.RowsAffected() != 1 {
		return ErrNotFound
	}

	return tx.Commit()
}

func (tx *Tx) getFamily(id uuid.UUID) (int, error) {
	var fam int
	err := tx.QueryRow("SELECT family FROM datasets WHERE id = $1", id.Array()).Scan(&fam)
	if err != nil {
		return 0, handleError(err)
	}

	return fam, nil
}

// CheckOwner returns an error if the record is not owned by the given user.
func (tx *Tx) CheckOwner(id uuid.UUID, owner uuid.UUID) error {
	var isOwner bool
	err := tx.QueryRow("SELECT (owner = $2) FROM datasets WHERE id = $1", id.Array(), owner.Array()).Scan(&isOwner)
	if err != nil {
		return handleError(err)
	}

	if !isOwner {
		return ErrNotOwner
	}

	return nil
}

// CheckOwner calls tx.CheckOwner to check if the record exists and is owner by the given user.
func (db *DB) CheckOwner(id uuid.UUID, owner uuid.UUID) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	return tx.CheckOwner(id, owner)
}

// Get retrieves a dataset from the database.
func (db *DB) Get(id uuid.UUID) (*models.Dataset, error) {
	var (
		valid  *bool
		family *int
		schema *string
		blob   []byte
	)

	res := new(models.Dataset)
	err := db.pool.QueryRow("select id, creator, owner, valid, family, schema, blob from datasets where id=$1", id.Array()).Scan(res.Id.Array(), res.Creator.Array(), res.Owner.Array(), &valid, &family, &schema, &blob)
	if err != nil {
		return nil, handleError(err)
	}

	err = res.SetData(*family, *schema, blob)
	if err != nil {
		return nil, err
	}

	res.SetValid(*valid)

	return res, nil
}

// GetWithOwner retrieves a dataset from the database if the owner matches.
func (db *DB) GetWithOwner(id uuid.UUID, owner uuid.UUID) (*models.Dataset, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	err = tx.CheckOwner(id, owner)
	if err != nil {
		return nil, err
	}

	return tx.get(id, "")
}

func (tx *Tx) get(id uuid.UUID, key string) (*models.Dataset, error) {
	var (
		family *int
		schema *string
		blob   []byte

		err error
	)

	res := new(models.Dataset)
	if key == "" {
		err = tx.QueryRow("select id, creator, owner, family, schema, blob from datasets where id=$1", id.Array()).Scan(res.Id.Array(), res.Creator.Array(), res.Owner.Array(), &family, &schema, &blob)
	} else {
		err = tx.QueryRow(`select id, creator, owner, family, schema, blob#>$2 from datasets where id=$1`, id.Array(), []string{key}).Scan(res.Id.Array(), res.Creator.Array(), res.Owner.Array(), &family, &schema, &blob)
	}
	if err != nil {
		return nil, handleError(err)
	}

	err = res.SetData(*family, *schema, blob)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Delete removes one dataset from the database if the owner matches.
func (db *DB) Delete(id uuid.UUID, owner *uuid.UUID) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if owner != nil {
		err = tx.CheckOwner(id, *owner)
		if err != nil {
			return handleError(err)
		}
	}

	ct, err := tx.Exec(`DELETE FROM datasets WHERE id = $1`, id.Array())
	if err != nil {
		return handleError(err)
	}

	if ct.RowsAffected() != 1 {
		return ErrNotFound
	}

	return tx.Commit()
}

// GetAllForUid returns all datasets for a given user.
func (db *DB) GetAllForUid(uid uuid.UUID) ([]*models.Dataset, error) {
	var list []*models.Dataset

	rows, err := db.pool.Query("select id, creator, owner, synced, family, schema, valid, blob from datasets where owner=$1", uid.Array())
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		var dataset models.Dataset
		var (
			synced *time.Time
			family int
			schema string
			valid  bool
			blob   []byte
		)
		err = rows.Scan(dataset.Id.Array(), dataset.Creator.Array(), dataset.Owner.Array(), &synced, &family, &schema, &valid, &blob)
		if err != nil {
			return nil, err
		}
		if synced != nil {
			dataset.Synced = *synced
		}
		dataset.SetData(family, schema, blob)
		if err != nil {
			return nil, err
		}
		list = append(list, &dataset)
	}

	if rows.Err() != nil {
		return []*models.Dataset{}, rows.Err()
	}

	return list, nil
}

// ListAllForUid returns the list of datasets for a given user.
func (db *DB) ListAllForUid(uid uuid.UUID) ([]*models.Dataset, error) {
	var list []*models.Dataset

	rows, err := db.pool.Query("select id, creator, owner, family, schema, valid from datasets where owner=$1", uid.Array())
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		var dataset models.Dataset
		var (
			family int
			schema string
			valid  bool
		)
		err = rows.Scan(dataset.Id.Array(), dataset.Creator.Array(), dataset.Owner.Array(), &family, &schema, &valid)
		if err != nil {
			return nil, err
		}
		dataset.SetData(family, schema, nil)
		if err != nil {
			return nil, err
		}
		list = append(list, &dataset)
	}

	if rows.Err() != nil {
		return []*models.Dataset{}, rows.Err()
	}

	return list, nil
}

// ChangeOwnerTo updates a dataset's owner.
func (db *DB) ChangeOwnerTo(id uuid.UUID, uid uuid.UUID) error {
	tx, err := db.pool.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tag, err := tx.Exec("UPDATE datasets SET owner = $1 WHERE id = $2", uid.Array(), id.Array())
	if err != nil {
		return handleError(err)
	}
	log.Println("tag:", tag)

	return tx.Commit()
}
