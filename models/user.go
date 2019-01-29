package models

import (
	"github.com/francoispqt/gojay"
	"github.com/wvh/uuid"
)

// User defines a basic user object to hold common session information.
type User struct {
	// Uid is the qvain user id.
	Uid uuid.UUID

	// Identity is the identity the user logged in with. Can be empty if the user didn't actually log in.
	Identity string

	// Service is the service from which the identity originated. Can be empty if the user didn't actually log in.
	Service string

	// Name is the user's name, if available.
	Name string

	// Email is the user's primary email address (if provided).
	Email string

	// Organisation is the organisation the external identity is a member of (if provided).
	Organisation string

	// Projects are a sort of user groups defined in the token.
	// This is a list instead of a map/set because lists are faster for small numbers of elements (~20).
	Projects []string
}

// HasProject returns a boolean value indicating whether a user is a member of a given project.
func (user *User) HasProject(project string) bool {
	for i := range user.Projects {
		if user.Projects[i] == project {
			return true
		}
	}
	return false
}

// MarshalJSONObject implements gojay.MarshalJSONObject.
func (user *User) MarshalJSONObject(enc *gojay.Encoder) {
	enc.StringKey("uid", user.Uid.String())
	enc.StringKey("identity", user.Identity)
	enc.StringKey("service", user.Service)
	enc.StringKey("name", user.Name)
	enc.StringKey("email", user.Email)
	enc.StringKey("organisation", user.Organisation)

	if len(user.Projects) > 0 {
		// ArrayKeyOmitEmpty only checks nil, not len
		enc.ArrayKey("projects", gojay.EncodeArrayFunc(func(enc *gojay.Encoder) {
			for i := range user.Projects {
				enc.AddString(user.Projects[i])
			}
		}))
	}
}

// IsNil implements gojay.IsNil interface.
func (u *User) IsNil() bool {
	return u == nil
}

// UnmarshalerJSONObject implements gojay.UnmarshalerJSONObject.
func (user *User) UnmarshalJSONObject(dec *gojay.Decoder, key string) (err error) {
	switch key {
	case "uid":
		var uid string
		dec.String(&uid)
		user.Uid, err = uuid.FromString(uid)
		if err != nil {
			return err
		}
	case "identity":
		return dec.String(&user.Identity)
	case "service":
		return dec.String(&user.Service)
	case "name":
		return dec.String(&user.Name)
	case "email":
		return dec.String(&user.Email)
	case "organisation":
		return dec.String(&user.Organisation)
	case "projects":
		var project string
		return dec.DecodeArray(gojay.DecodeArrayFunc(func(dec *gojay.Decoder) error {
			if err := dec.String(&project); err != nil {
				return err
			}
			user.Projects = append(user.Projects, project)
			return nil
		}))
	}
	return nil
}
func (u *User) NKeys() int {
	return 7
}

// UnmarshalJSON implements the Unmarshaler interface from the standard library json package.
func (user *User) UnmarshalJSON(b []byte) error {
	return gojay.Unmarshal(b, user)
}

// UserFromJson constructs a new user object by calling unmarshal on the given json byte slice.
func UserFromJson(b []byte) (u *User, err error) {
	u = new(User)
	err = gojay.UnmarshalJSONObject(b, u)
	return
}
