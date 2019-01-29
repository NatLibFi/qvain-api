package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/wvh/uuid"
)

var (
	ErrIdMissing          = errors.New("missing dataset id")
	ErrIdPresent          = errors.New("dataset already has id")
	ErrDatasetTypeMissing = errors.New("dataset type is missing")
	ErrSchemaMissing      = errors.New("schema identifier is missing")
	ErrDatasetMissing     = errors.New("dataset is missing or empty")
)

// TypedDataset is a wrapper around a base dataset that allows different dataset types to fondle the data in ways that pleases them.
type TypedDataset interface {
	CreateData(int, string, []byte, map[string]string) error
	UpdateData(int, string, []byte, map[string]string) error
	Unwrap() *Dataset
}

// CreateDatasetFromJson creates a typed dataset with JSON data originating from the web API.
// Caller closes body.
func CreateDatasetFromJson(creator uuid.UUID, data io.Reader, inject map[string]string) (TypedDataset, error) {
	decoder := json.NewDecoder(data)

	// partially decode json
	aux := &struct {
		Id *uuid.UUID `json:"id"`

		Family *int             `json:"type"`
		Schema *string          `json:"schema"`
		Blob   *json.RawMessage `json:"dataset"`

		Valid bool `json:"valid"`
	}{}
	err := decoder.Decode(aux)
	if err != nil {
		return nil, err
	}

	if aux.Id != nil {
		return nil, ErrIdPresent
	}

	if aux.Family == nil {
		return nil, ErrDatasetTypeMissing
	}

	if aux.Schema == nil {
		return nil, ErrSchemaMissing
	}

	if aux.Blob == nil || len(*aux.Blob) == 0 {
		return nil, ErrDatasetMissing
	}

	fam, err := LookupFamily(*aux.Family)
	if err != nil {
		return nil, err
	}

	typed, err := fam.NewFunc(creator)
	if err != nil {
		return nil, err
	}

	err = typed.CreateData(*aux.Family, *aux.Schema, *aux.Blob, inject)
	if err != nil {
		return nil, err
	}

	return typed, nil
}

// UpdateDatasetFromJson makes a – potentially partial – typed dataset with JSON data originating from the web API.
//func UpdateDatasetFromJson(owner uuid.UUID, data []byte) (TypedDataset, error) {
func UpdateDatasetFromJson(owner uuid.UUID, data io.Reader, inject map[string]string) (TypedDataset, error) {
	decoder := json.NewDecoder(data)

	aux := &struct {
		Id *uuid.UUID `json:"id"`

		Family int              `json:"type"`
		Schema string           `json:"schema"`
		Blob   *json.RawMessage `json:"dataset"`

		Valid bool `json:"valid"`
	}{}
	err := decoder.Decode(aux)
	//err := json.Unmarshal(data, aux)
	if err != nil {
		return nil, err
	}

	if aux.Id == nil {
		return nil, ErrIdMissing
	}

	fam, err := LookupFamily(aux.Family)
	if err != nil {
		return nil, ErrInvalidFamily
	}

	typed := fam.LoadFunc(&Dataset{
		Id:    *aux.Id,
		Owner: owner,
	})

	fmt.Printf("typed: %#v\n", typed)

	err = typed.UpdateData(aux.Family, aux.Schema, *aux.Blob, inject)
	if err != nil {
		return nil, err
	}

	return typed, nil
}
