package sessions

import (
	"net/http"
	"strings"
)

func getBearerToken(r *http.Request) string {
	hdr := r.Header.Get("Authorization")
	if len(hdr) > 7 && hdr[0:7] == "Bearer " {
		return hdr[7:]
	}
	return ""
}

// GetJwtSignature returns the signature of JWT tokens.
// It is meant for HS* and RS* symmetrical and elliptical algorithms only with an encoded length of 186B;
// it is up to the caller to make sure this only gets called with relevant signatures. Here be dragons.
func GetJwtSignature(jwt string) (string, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return "", ErrMalformedToken
	}
	// HS256 = 32B/43, HS512 = 64B/86, RS256 = ~64B/86; max for curves ~139B/186
	// see: base64.URLEncoding.WithPadding(base64.NoPadding).EncodedLen()
	if len(parts[2]) < 40 || len(parts[2]) > 186 {
		return "", ErrMalformedToken
	}
	return "jwt:" + parts[2], nil
}
