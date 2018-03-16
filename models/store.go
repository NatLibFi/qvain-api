
package models

import (
	"wvh/uuid"
)


type Store struct {}

func NewStore() {}

func (store *Store) Save(blob string) error {
	return nil
}

func (store *Store) Retrieve(id uuid.UUID) (string, error) {
	return "", nil
}

