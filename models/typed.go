package models

import (
	"encoding/json"
	"errors"

	"github.com/wvh/uuid"
)

var (
	ErrInvalidDatasetType = errors.New("invalid dataset type")
	ErrIdMissing          = errors.New("missing dataset id")
	ErrIdPresent          = errors.New("dataset already has id")
)

// TypedDataset is a wrapper around a base dataset that allows different dataset types to fondle the data in ways that pleases them.
type TypedDataset interface {
	CreateData(int, string, []byte) error
	UpdateData(int, string, []byte) error
	Unwrap() *Dataset
}

// CreateDatasetFromJson creates a typed dataset with JSON data originating from the web API.
func CreateDatasetFromJson(creator uuid.UUID, data []byte) (TypedDataset, error) {
	// partially decode json
	aux := &struct {
		Id *uuid.UUID `json:"id"`

		Family int              `json:"family"`
		Schema string           `json:"schema"`
		Blob   *json.RawMessage `json:"dataset"`

		Valid bool `json:"valid"`
	}{}
	err := json.Unmarshal(data, aux)
	if err != nil {
		return nil, err
	}

	if aux.Id != nil {
		return nil, ErrIdPresent
	}

	fam := LookupFamily(aux.Family)
	if fam == nil {
		return nil, ErrInvalidDatasetType
	}

	typed, err := fam.NewFunc(creator)
	if err != nil {
		return nil, err
	}

	err = typed.CreateData(aux.Family, aux.Schema, *aux.Blob)
	if err != nil {
		return nil, err
	}

	return typed, nil
}

// UpdateDatasetFromJson makes a – potentially partial – typed dataset with JSON data originating from the web API.
func UpdateDatasetFromJson(owner uuid.UUID, data []byte) (TypedDataset, error) {
	aux := &struct {
		Id *uuid.UUID `json:"id"`

		Family int              `json:"family"`
		Schema string           `json:"schema"`
		Blob   *json.RawMessage `json:"dataset"`

		Valid bool `json:"valid"`
	}{}
	err := json.Unmarshal(data, aux)
	if err != nil {
		return nil, err
	}

	if aux.Id == nil {
		return nil, ErrIdMissing
	}

	fam := LookupFamily(aux.Family)
	if fam == nil {
		return nil, ErrInvalidDatasetType
	}

	typed := fam.LoadFunc(&Dataset{
		Id:    *aux.Id,
		Owner: owner,
	})

	err = typed.UpdateData(aux.Family, aux.Schema, *aux.Blob)
	if err != nil {
		return nil, err
	}

	return typed, nil
}
