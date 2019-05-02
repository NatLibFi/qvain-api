package models

import (
	"errors"
	"time"

	"github.com/wvh/uuid"
)

var (
	errNeedFamily  = errors.New("need schema type for dataset")
	errNeedSchema  = errors.New("need schema name for dataset")
	errNeedDataset = errors.New("need dataset blob")
)

type Dataset struct {
	Id      uuid.UUID
	Creator uuid.UUID
	Owner   uuid.UUID

	Created  time.Time
	Modified time.Time
	Synced   time.Time

	Published bool
	valid     bool

	family int
	schema string
	blob   []byte
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

// SetData sets the schema family and name as well as the metadata blob.
// It is an error not to provide the appropriate schema family and name.
func (ds *Dataset) SetData(family int, schema string, blob []byte) error {
	if family < 0 {
		return errNeedFamily
	}
	if schema == "" {
		return errNeedSchema
	}
	if blob == nil {
		blob = []byte("")
	}
	ds.family = family
	ds.schema = schema
	ds.blob = blob

	return nil
}

// CreateData allows dataset types to override what happens on dataset creation.
func (ds *Dataset) CreateData(family int, schema string, blob []byte, extra map[string]string) error {
	return ds.SetData(family, schema, blob)
}

// UpdateData allows dataset types to override what happens on update of an existing dataset.
func (ds *Dataset) UpdateData(family int, schema string, blob []byte, extra map[string]string) error {
	return ds.SetData(family, schema, blob)
}

func (ds *Dataset) Family() int {
	return ds.family
}

func (ds *Dataset) Schema() string {
	return ds.schema
}

func (ds *Dataset) Blob() []byte {
	return ds.blob
}

func (ds *Dataset) SetValid(valid bool) {
	ds.valid = valid
}

func (ds *Dataset) IsValid() bool {
	return ds.valid
}

func (ds *Dataset) Unwrap() *Dataset {
	return ds
}
