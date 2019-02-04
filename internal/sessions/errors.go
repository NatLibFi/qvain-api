package sessions

import (
	"errors"
)

var (
	// ErrSessionNotFound is the error returned if we can't find a given session ID or no session ID has been given at all.
	ErrSessionNotFound = errors.New("expired or invalid session")

	// ErrUnknownUser occurs when the user successfully logs in with an external identity but we can't get our uid from the database.
	ErrUnknownUser = errors.New("unknown user")

	// ErrCreatingSid occurs when we can't read some random bytes from the system; this error is highly improbable.
	ErrCreatingSid = errors.New("error creating random session id")

	// ErrMalformedToken happens when we can't parse a bearer token.
	ErrMalformedToken = errors.New("malformed token")

	// ErrTokenConfigError is returned when trying to create a new session from token without having a token func defined.
	ErrTokenConfigError = errors.New("token session without token configuration")
)
