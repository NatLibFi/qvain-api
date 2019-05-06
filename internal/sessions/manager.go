// Package session implements a simple session manager for token authentication.
//
// This is not meant to be a generic session library, but to be used specifically as part of the software it comes with.
package sessions

import (
	"fmt"
	//"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/CSCfi/qvain-api/internal/randomkey"
	"github.com/CSCfi/qvain-api/pkg/models"
	"github.com/francoispqt/gojay"
	"github.com/gomodule/redigo/redis"
	"github.com/muesli/cache2go"
	"github.com/wvh/uuid"
)

// DefaultExpiration is the default duration before a session expires
const DefaultExpiration = 2 * 60 * time.Minute

type sidGenerator func(string) (string, error)

// Manager handles the actual storage and retrieval of sessions.
type Manager struct {
	cache              *cache2go.CacheTable
	onToken            func(string) (string, error)
	genTokenSid        sidGenerator
	RequireCSCUserName bool
}

type ManagerOption func(*Manager)

func WithRequireCSCUserName(require bool) ManagerOption {
	return func(mgr *Manager) {
		mgr.RequireCSCUserName = require
	}
}

// NewManager creates a new session storage.
func NewManager(opts ...ManagerOption) *Manager {
	mgr := &Manager{
		cache: cache2go.Cache("sessions"),
	}
	for _, opt := range opts {
		opt(mgr)
	}
	return mgr
}

// SetOnToken takes a function that can create a session from a token, and optionally
// a second function that can securily shorten a token to generate a session identifier.
// This function is not safe to run after the session manager has been taken into use.
func (mgr *Manager) SetOnToken(on func(string) (string, error), gen sidGenerator) {
	mgr.onToken = on
	mgr.genTokenSid = gen
}

// new creates the actual session with a session id either from a random hash or calculated from a given token.
func (mgr *Manager) new(sid string, uid *uuid.UUID, user *models.User, opts ...SessionOption) error {
	session := &Session{
		uid:        uid,
		User:       user,
		Expiration: time.Now().Add(DefaultExpiration),
	}
	for _, f := range opts {
		f(session)
	}

	mgr.cache.Add(sid, DefaultExpiration, session)
	return nil
}

// NewFromToken creates a session from a token. The session manager needs to have been configured for tokens by SetOnToken().
func (mgr *Manager) NewFromToken(token string, uid *uuid.UUID, user *models.User, opts ...SessionOption) error {
	// don't allow token sessions without having a token func defined
	if mgr.onToken == nil {
		return ErrTokenConfigError
	}

	var (
		sid string
		err error
	)

	// generate sid for token
	if mgr.genTokenSid != nil {
		sid, err = mgr.genTokenSid(token)
		if err != nil {
			return err
		}
	} else {
		sid = "token:" + token
	}

	return mgr.new(sid, uid, user, opts...)
}

// NewLogin logs in a user by creating a session.
func (mgr *Manager) NewLogin(uid *uuid.UUID, user *models.User, opts ...SessionOption) (string, error) {
	key, err := randomkey.Random16()
	if err != nil {
		return "", ErrCreatingSid
	}
	sid := key.Base64()

	return sid, mgr.new(sid, uid, user, opts...)
}

// NewLoginWithCookie wraps NewLogin to set a session cookie.
func (mgr *Manager) NewLoginWithCookie(w http.ResponseWriter, uid *uuid.UUID, user *models.User, opts ...SessionOption) (string, error) {
	sid, err := mgr.NewLogin(uid, user, opts...)
	if err != nil {
		return sid, err
	}
	SetSessionCookie(w, sid)
	return sid, err
}

func (mgr *Manager) Get(sid string) (*Session, error) {
	s, err := mgr.cache.Value(sid)
	if err != nil {
		if err == cache2go.ErrKeyNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	return s.Data().(*Session), nil
}

func (mgr *Manager) Exists(sid string) bool {
	return mgr.cache.Exists(sid)
}

func (mgr *Manager) Destroy(sid string) bool {
	_, err := mgr.cache.Delete(sid)
	//if err != nil && err != cache2go.ErrKeyNotFound {
	if err != nil {
		return false
	}
	return true
}

func (mgr *Manager) DestroyWithCookie(w http.ResponseWriter, sid string) bool {
	DeleteSessionCookie(w)
	return mgr.Destroy(sid)
}

func (mgr *Manager) Count() int {
	return mgr.cache.Count()
}

func (mgr *Manager) List(w io.Writer) {
	enc := gojay.NewEncoder(w)
	defer enc.Release()

	enc.EncodeArray(gojay.EncodeArrayFunc(func(enc *gojay.Encoder) {
		mgr.cache.Foreach(func(key interface{}, item *cache2go.CacheItem) {
			session := item.Data().(*Session)
			enc.AddObject(session)
		})
	}))
}

func (mgr *Manager) Save() {
	conn, err := redis.Dial("unix", "/home/wouter/.q.redis.sock")
	if err != nil {
		panic(err)
	}
	_ = conn
	exp := int64(DefaultExpiration / time.Second)

	mgr.cache.Foreach(func(key interface{}, item *cache2go.CacheItem) {
		fmt.Printf("item: %+v\n", item)
		session := item.Data().(*Session)
		json, err := gojay.MarshalJSONObject(session)
		if err != nil {
			panic(err)
		}
		//strKey := key.(string)
		fmt.Println("writing to redis:", key, exp, json)
		conn.Send("SETEX", key, exp, json)
		conn.Flush()
	})
	conn.Flush()
}

// SessionFromRequest returns the existing session for the request or, failing that, an error.
func (mgr *Manager) SessionFromRequest(r *http.Request) (*Session, error) {
	// get cookie
	sid, err := GetSessionCookie(r)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	// login with cookie
	if sid != "" {
		return mgr.Get(sid)
	}

	// if tokens are enabled...
	if mgr.onToken != nil {
		// get token from header
		token := getBearerToken(r)
		if token == "" {
			return nil, ErrSessionNotFound
		}

		// generate sid for token
		if mgr.genTokenSid != nil {
			sid, err = mgr.genTokenSid(token)
			if err != nil {
				return nil, err
			}
		} else {
			sid = "token:" + token
		}

		// check the cache if the token has "logged in"
		session, err := mgr.Get(sid)
		if err == nil {
			return session, nil
		}

		// login with token callback
		sid, err = mgr.onToken(token)
		if err != nil {
			return nil, err
		}

		return mgr.Get(sid)
	}

	return nil, ErrSessionNotFound
}

// UserSessionFromRequest gets the session for the current request if one exists and checks if it has a valid user.
// This is a shortcut that calls SessionFromRequest followed by HasUser.
func (mgr *Manager) UserSessionFromRequest(r *http.Request) (*Session, error) {
	session, err := mgr.SessionFromRequest(r)
	if err != nil {
		return nil, err
	}
	if session.HasUser() {
		return session, nil
	}
	return nil, ErrUnknownUser
}

func (mgr *Manager) GetRedis(conn redis.Conn, key string) ([]byte, error) {
	res, err := conn.Do("GET", key)
	if err != nil {
		return []byte(""), err
	}
	return res.([]byte), nil
	//return redis.String(conn.Do("GET", key))
}

type SessionOption func(*Session)

func WithExpiration(expAt time.Time) SessionOption {
	return func(session *Session) {
		session.Expiration = expAt
	}
}

func WithDuration(exp time.Duration) SessionOption {
	return func(session *Session) {
		session.Expiration = time.Now().Add(exp)
	}
}
