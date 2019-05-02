package psql

import (
	"time"

	"github.com/CSCfi/qvain-api/pkg/models"
	"github.com/wvh/uuid"
)

type BatchManager struct {
	tx         *Tx
	triggerUid *uuid.UUID
	at         time.Time
}

func (db *DB) NewBatch() (*BatchManager, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	return &BatchManager{tx: tx, at: time.Now()}, nil
}

func (db *DB) NewBatchForUser(uid uuid.UUID) (*BatchManager, error) {
	b, err := db.NewBatch()
	if err != nil {
		return nil, err
	}

	b.triggerUid = &uid
	return b, nil
}

func (b *BatchManager) Create(dataset *models.Dataset) error {
	return b.tx.Create(dataset)
}

func (b *BatchManager) CreateWithMetadata(dataset *models.Dataset) error {
	dataset.Synced = b.at
	return b.tx.createWithMetadata(dataset)
}

func (b *BatchManager) UpdateSynced(id uuid.UUID) error {
	return b.tx.updateSyncedByService(id)
}

func (b *BatchManager) Update(id uuid.UUID, blob []byte) error {
	return b.tx.updateByService(id, blob)
}

func (b *BatchManager) Upsert(data *models.Dataset) error {
	return ErrNotImplemented
}

func (b *BatchManager) writeStamp() error {
	if b.triggerUid == nil {
		return nil
	}
	_, err := b.tx.Exec(`INSERT INTO lastsync(uid, ts, success) VALUES($1, $2, $3) ON CONFLICT (uid) DO UPDATE SET ts = $2, success = $3 WHERE lastsync.uid = $1`, b.triggerUid.Array(), time.Now(), true)
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
	err := tx.QueryRow("SELECT ts FROM lastsync WHERE uid = $1", uid.Array()).Scan(&last)
	if err != nil {
		return time.Time{}, handleError(err)
	}

	return last, nil
}
