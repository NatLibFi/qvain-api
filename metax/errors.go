package metax

import (
	"encoding/json"
	"errors"
)

var (
	ErrInvalidContentType = errors.New("invalid content-type: expected json")
	ErrNotFound           = errors.New("not found")
	ErrIdRequired         = errors.New("dataset without id and not allowed to create")
	ErrInvalidId          = errors.New("invalid dataset id")
)

// LinkingError is a custom error type that adds the missing field name.
type LinkingError struct {
	field string
}

// NewLinkingError creates a new LinkingError with the field name set to the argument, or marks the whole metadata block as missing if no argument given.
func NewLinkingError(field ...string) *LinkingError {
	if len(field) < 1 {
		return &LinkingError{}
	}
	return &LinkingError{field: field[0]}
}

// Error satisfies the Error interface.
func (e *LinkingError) Error() string {
	if e.field == "" {
		return "no qvain metadata"
	}
	return "missing field: " + e.field
}

// Field returns the missing field name.
func (e *LinkingError) Field() string {
	return e.field
}

// IsNotMine returns a boolean value indicating if the whole Qvain metadata block was missing.
func (e *LinkingError) IsNotMine() bool {
	return e.field == ""
}

type ApiError struct {
	myError    string
	metaxError json.RawMessage
	statusCode int
}

func (e *ApiError) Error() string {
	return e.myError
}

func (e *ApiError) StatusCode() int {
	return e.statusCode
}

func (e *ApiError) OriginalError() json.RawMessage {
	return e.metaxError
}
