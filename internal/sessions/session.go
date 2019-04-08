package sessions

import (
	"time"

	"github.com/CSCfi/qvain-api/pkg/models"
	"github.com/francoispqt/gojay"
	"github.com/wvh/uuid"
)

var (
	// null is a typed json null.
	null = gojay.EmbeddedJSON("null")
)

// Session contains user session information.
type Session struct {
	// uid is our own user id.
	uid *uuid.UUID

	// Expiration is the maximum expiration time for the session.
	Expiration time.Time

	// User is the application user object.
	User *models.User
}

// Uid returns the user id or an error if the session doesn't have a valid (application) user.
func (session *Session) Uid() (uuid.UUID, error) {
	if session.uid == nil {
		return uuid.UUID{}, ErrUnknownUser
	}
	return *session.uid, nil
}

// MaybeUid is a convenience function that returns the user id as a string or empty string if not set.
func (session *Session) MaybeUid() string {
	if session.uid != nil {
		return session.uid.String()
	}
	return ""
}

// HasUser returns true if the session is an end user session with a valid user object.
func (session *Session) HasUser() bool {
	return session.User != nil
}

func (session *Session) MarshalJSONObject(enc *gojay.Encoder) {
	if session.uid != nil {
		enc.StringKey("uid", session.uid.String())
	} else {
		enc.AddEmbeddedJSONKey("uid", &null)
	}

	enc.Int64Key("expiration", session.Expiration.Unix())

	// if want null rather than omit:
	//   enc.ObjectKeyNullEmpty("user", session.User)
	enc.ObjectKeyOmitEmpty("user", session.User)
}

func (session *Session) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "uid":
		var uidString string
		dec.String(&uidString)
		if uidString == "" {
			return nil
		} else if uid, err := uuid.FromString(uidString); err != nil {
			return err
		} else {
			session.uid = &uid
		}
	case "expiration":
		var epoch int64
		err := dec.Int64(&epoch)
		if err != nil {
			return err
		}
		session.Expiration = time.Unix(epoch, 0)
	case "user":
		// if JSON null should also be allowed:
		//   return dec.ObjectNull(&session.User)
		session.User = new(models.User)
		return dec.Object(session.User)
	}
	return nil
}

func (session *Session) NKeys() int {
	return 3
}

// IsNil returns a boolean indicating whether the session is nil (method required by gojay JSON library).
func (session *Session) IsNil() bool {
	return session == nil
}

// AsJson returns a byte slice containing the JSON representation of the session.
// See also the Public() method that removes private information.
func (session *Session) AsJson() ([]byte, error) {
	return gojay.MarshalJSONObject(session)
}

// FromJson returns a session deserialised from its JSON representation.
func FromJson(json []byte, session *Session) error {
	return gojay.UnmarshalJSONObject(json, session)
}

// Public makes a chainable copy of a session and removes fields that should not be shown to the outside world.
func (session Session) Public() *Session {
	public := session

	// we don't actually have any fields that are not public right now...
	//public.Projects = []string{}

	return &public
}
