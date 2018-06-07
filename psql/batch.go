package psql

import (
	"time"

	"github.com/NatLibFi/qvain-api/models"
	"github.com/wvh/uuid"
)

type BatchManager struct {
	tx         *Tx
	triggerUid *uuid.UUID
}

func (db *DB) NewBatch() (*BatchManager, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	//defer tx.Rollback()
	return &BatchManager{tx: tx}, nil
}

func (db *DB) NewBatchWithUser(uid uuid.UUID) (*BatchManager, error) {
	b, err := db.NewBatch()
	if err != nil {
		return nil, err
	}

	b.triggerUid = &uid
	return b, nil
}

func (b *BatchManager) Store(dataset *models.Dataset) error {
	return b.tx.Store(dataset)
}

func (b *BatchManager) writeStamp() error {
	if b.triggerUid == nil {
		return nil
	}
	_, err := b.tx.Exec(`INSERT INTO lastsync(uid, ts, success) VALUES($1, $2, $3) ON CONFLICT DO UPDATE SET ts = $2, success = $3 WHERE uid = $1`, b.triggerUid.Array(), time.Now(), true)
	return err
}

func (b *BatchManager) Commit() error {
	err := b.writeStamp()
	if err != nil {
		return err
	}
	return b.tx.Commit()
}

func (b *BatchManager) Rollback() {
	defer b.tx.Rollback()
}

func (db *DB) GetLastSync(uid uuid.UUID) (time.Time, error) {
	tx, err := db.Begin()
	if err != nil {
		return time.Time{}, err
	}
	defer tx.Rollback()
	return tx.getLastSync(uid)
}

func (tx *Tx) getLastSync(uid uuid.UUID) (time.Time, error) {
	var last time.Time
	err := tx.QueryRow("SELECT ts FROM lastsync WHERE id = $1", uid.Array()).Scan(&last)
	if err != nil {
		return time.Time{}, handleError(err)
	}

	return last, nil
}
