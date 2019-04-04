package models

import (
	"errors"

	"github.com/wvh/uuid"
)

var ErrInvalidFamily = errors.New("Invalid dataset type")

var privateTypeRegistry *TypeRegistry

type TypeRegistry struct {
	tmap map[int]*SchemaFamily
}

func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{tmap: make(map[int]*SchemaFamily)}
}

func (reg *TypeRegistry) Register(id int, name string, newFunc NewFunc, loadFunc LoadFunc, paths []string) {
	reg.tmap[id] = &SchemaFamily{
		Id:          id,
		Name:        name,
		NewFunc:     newFunc,
		LoadFunc:    loadFunc,
		publicPaths: paths,
	}
}

func (reg *TypeRegistry) Lookup(id int) (*SchemaFamily, error) {
	if t, e := reg.tmap[id]; e {
		return t, nil
	}
	return nil, ErrInvalidFamily
}

func init() {
	// global registry
	privateTypeRegistry = NewTypeRegistry()

	privateTypeRegistry.Register(0, "no type", NewUntypedDataset, LoadUntypedDataset, nil)
	privateTypeRegistry.Register(1, "open dataset", NewOpenDataset, LoadOpenDataset, nil)
	// metax registers its own dataset type(s)
}

// RegisterFamily registers a dataset type into the global registry.
func RegisterFamily(id int, name string, newFunc NewFunc, loadFunc LoadFunc, paths []string) {
	privateTypeRegistry.Register(id, name, newFunc, loadFunc, paths)
}

// LookupFamily looks up a dataset type from the global registry.
func LookupFamily(id int) (*SchemaFamily, error) {
	return privateTypeRegistry.Lookup(id)
}

// Let's pre-define some basic dataset types... Not sure if this is the best place for it.

// UntypedDataset is a fall-back dataset type for datasets without type.
type UntypedDataset struct {
	*Dataset
}

// NewUntypedDataset creates an untyped dataset.
func NewUntypedDataset(creator uuid.UUID) (TypedDataset, error) {
	ds, err := NewDataset(creator)
	if err != nil {
		return nil, err
	}

	return &UntypedDataset{Dataset: ds}, nil
}

// LoadUntypedDataset constructs an existing untyped dataset from an existing base dataset.
func LoadUntypedDataset(ds *Dataset) TypedDataset {
	return &UntypedDataset{Dataset: ds}
}

// OpenDataset is a dataset type with no restrictions.
type OpenDataset struct {
	*Dataset
}

// NewOpenDataset creates a new Open (unrestricted) dataset.
func NewOpenDataset(creator uuid.UUID) (TypedDataset, error) {
	ds, err := NewDataset(creator)
	if err != nil {
		return nil, err
	}

	return &OpenDataset{Dataset: ds}, nil
}

// LoadUntypedDataset constructs an existing open dataset from an existing base dataset.
func LoadOpenDataset(ds *Dataset) TypedDataset {
	return &OpenDataset{Dataset: ds}
}
