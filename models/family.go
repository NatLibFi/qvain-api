package models

import (
	"strings"

	"github.com/wvh/uuid"
)

// NewFunc is a constructor function that creates a new typed datatset satisfying the TypedDataset interface.
type NewFunc func(uuid.UUID) (TypedDataset, error)

// LoadFunc is a constructor function that wraps a base dataset returning a typed dataset satisfying the TypedDataset interface.
type LoadFunc func(*Dataset) TypedDataset

// SchemaFamily defines a dataset type.
type SchemaFamily struct {
	Id          int
	Name        string
	NewFunc     NewFunc
	LoadFunc    LoadFunc
	publicPaths []string
}

// IsPartial returns a boolean indicating if this is a dataset that is only partially shown to the API.
func (fam *SchemaFamily) IsPartial() bool {
	// nil slice also has len 0
	return len(fam.publicPaths) > 0
}

// Key returns the key containing the partial dataset; the first element of the publicPaths slice.
func (fam *SchemaFamily) Key() string {
	if len(fam.publicPaths) > 0 {
		return fam.publicPaths[0]
	}
	return ""
}

// IsPathPublic returns a boolean indicating if the dataset's subkey can be shown via API.
// A nil path list means no restrictions.
func (fam *SchemaFamily) IsPathPublic(p string) bool {
	if fam.publicPaths == nil {
		return true
	}
	return inPrefixes(fam.publicPaths, p)
}

// contains does a simple linear string search.
// Slices are faster than maps for small collections (< 5 elements); as we don't expect to exceed 5 sub-paths, don't create a map.
func contains(a []string, s string) bool {
	for _, v := range a {
		if s == v {
			return true
		}
	}
	return false
}

// inPrefixes does a simple linear string prefix search.
func inPrefixes(prefixes []string, s string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
