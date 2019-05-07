package sessions

import (
	"net/http"
	"time"
)

const (
	// SessionCookieName is the name of the session cookie.
	SessionCookieName = "sid"

	// SessionCookiePath is the path for which the session cookie is valid.
	SessionCookiePath = "/"
)

// SetSessionCookie writes the session id cookie to the response.
func SetSessionCookie(w http.ResponseWriter, sid string) {
	http.SetCookie(w, &http.Cookie{
		Name:  SessionCookieName,
		Value: sid,
		Path:  SessionCookiePath,
		//MaxAge: int(DefaultExpiration / time.Second),
		// for old browsers; if not set, this will be a "session cookie"
		//Expires:  time.Now().Add(DefaultExpiration),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetSessionCookie tries to retrieve a session cookie from the request.
func GetSessionCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// DeleteSessionCookie tries to delete the session cookie from the browser.
func DeleteSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:  SessionCookieName,
		Value: "",
		Path:  SessionCookiePath,
		// doc: "MaxAge < 0 means delete cookie"
		MaxAge: -1,
		// for old browsers, send epoch
		Expires:  time.Unix(0, 0),
		Secure:   true,
		HttpOnly: true,
	})
}
