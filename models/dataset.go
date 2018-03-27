package models

import (
	"errors"
	"github.com/wvh/uuid"
	"time"
)

var (
	errNeedFamily = errors.New("need schema family type for metadata")
	errNeedSchema = errors.New("need schema name for metadata")
)

type Dataset struct {
	Id      uuid.UUID
	Creator uuid.UUID
	Owner   uuid.UUID

	Created  time.Time
	Modified time.Time

	Pushed time.Time
	Pulled time.Time

	family int
	schema string
	blob   string

	valid bool
}

// NewDataset creates a new dataset record with given Creator (which is also set into the Owner field).
// NOTE: the database will set Created and Modified dates to Now() by default.
func NewDataset(creator uuid.UUID) (*Dataset, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	//now := time.Now()

	return &Dataset{
		Id:      id,
		Creator: creator,
		Owner:   creator,
	}, nil
}

// SetMetadata sets the schema family and name as well as the metadata blob.
// It is an error not to provide the appropriate schema family and name.
func (ds *Dataset) SetMetadata(family int, schema, blob string) error {
	if family == 0 {
		return errNeedFamily
	}
	if schema == "" {
		return errNeedSchema
	}
	ds.family = family
	ds.schema = schema
	ds.blob = blob

	return nil
}

func (ds *Dataset) Family() int {
	return ds.family
}

func (ds *Dataset) Schema() string {
	return ds.schema
}

func (ds *Dataset) Blob() string {
	return ds.blob
}

func (ds *Dataset) SetValid(valid bool) {
	ds.valid = valid
}

func (ds *Dataset) IsValid() bool {
	return ds.valid
}
