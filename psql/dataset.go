
package psql

import (
	//"errors"

	"github.com/wvh/uuid"
	"github.com/NatLibFi/qvain-api/models"	
	"log"
	//"github.com/jackc/pgx"
)


func (db *PsqlService) ChangeOwnerTo(id uuid.UUID, uid uuid.UUID) error {
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
	
	err = tx.Commit()
	if err != nil {
		return err
	}
	
	return nil
}


func (db *PsqlService) Store(dataset *models.Dataset) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	err = tx.Store(dataset)
	if err != nil {
		return handleError(err)
	}
	
	err = tx.Commit()
	if err != nil {
		return err
	}
	
	return nil
}


func (db *PsqlService) BatchStore(datasets []*models.Dataset) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// do something batch-like
	for _, dataset := range datasets {
		err = tx.Store(dataset)
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


func (tx *Tx) Store(dataset *models.Dataset) error {
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


func (db *PsqlService) Get(id uuid.UUID) (*models.Dataset, error) {
	var (
		family *int
		schema *string
		blob   *string
	)
	res := new(models.Dataset)
	err := db.pool.QueryRow("select id, creator, owner, family, schema, blob from datasets where id=$1", id.Array()).Scan(res.Id, res.Creator, res.Owner, family, schema, blob)
	if err != nil {
		return nil, err
	}
	res.SetMetadata(*family, *schema, *blob)
	return res, nil
}


func (db *PsqlService) ListAllForUid(uid uuid.UUID) ([]*models.Dataset, error) {
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
			valid bool
		)
		err = rows.Scan(dataset.Id, dataset.Creator, dataset.Owner, family, schema, valid)
		if err != nil {
			return nil, err
		}
		dataset.SetMetadata(family, schema, "")
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
