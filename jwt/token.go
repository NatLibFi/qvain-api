// Package jwt implements a simple JWT middleware handler.
package jwt

import (
	"log"
	//"time"
	"errors"
	"net/http"
	
	"github.com/lestrrat/go-jwx/jwt"
	//"github.com/lestrrat/go-jwx/jws"
	"github.com/lestrrat/go-jwx/jwa"
)


// Default signature algorithm as we ignore the JOSE header.
const DefaultSignAlgorithm = jwa.HS256


var (
	errNeedSecretKey = errors.New("need secret key to verify JWT tokens")
	errInvalidToken  = errors.New("invalid token")
	errMissingJwtID  = errors.New("mandatory field `jti` is missing")
)


type opt func(*JwtHandler)


// Verbose is an option to show the actual validation error instead of a generic (but more secure) invalid token message.
func Verbose(jh *JwtHandler) {
	jh.verbose = true
}


// RequireJwtID is an option to require presence of the `jti` field to pass claim validation.
func RequireJwtID(jh *JwtHandler) {
	jh.mustJwtID = true
}


// WithErrorFunc is an option to set the function called in case of error. It follows the signature of http.Error(), which it also defaults to.
func WithErrorFunc(fn func(http.ResponseWriter, string, int)) opt {
	return func(jh *JwtHandler) {
		jh.errorFunc = fn
	}
}


// WithSigningAlgorithm allows setting the signing algorithm for the token.
// It takes one of the algorithms defined in the jwx/jwa sub-package.
func WithSigningAlgorithm(algo jwa.SignatureAlgorithm) opt {
	return func(jh *JwtHandler) {
		jh.signAlgorithm = algo
	}
}


// JwtHandler contains the configuration options for the JWT middleware handler.
type JwtHandler struct {
	audience  string
	key       []byte
	verbose   bool
	mustJwtID bool

	// see jwx/jwa/signature.go for options
	signAlgorithm jwa.SignatureAlgorithm
	
	// http.Error() signature
	errorFunc func(http.ResponseWriter, string, int)
}


// NewJwtHandler configures a JWT token middleware handler.
func NewJwtHandler(key []byte, audience string, opts... opt) *JwtHandler {
	if len(key) < 1 {
		panic(errNeedSecretKey)
	}
	
	jh := &JwtHandler{
		key:           key,
		audience:      audience,
		signAlgorithm: DefaultSignAlgorithm,
		errorFunc:     http.Error,
	}
	
	for _, f := range opts {
		f(jh)
	}
	
	return jh
}


// TokenError calls the set ErrorFunc function while taking the level of verbosity into consideration.
func (jh *JwtHandler) TokenError(w http.ResponseWriter, err string, code int) {
	if jh.verbose {
		jh.errorFunc(w, errInvalidToken.Error() + ": " + err, code)
	} else {
		jh.errorFunc(w, errInvalidToken.Error(), code)
	}
}


// TokenFromHeader retrieves a bearer token from Authorization header.
//
// Header format: "Bearer" 1*SP b64token (https://tools.ietf.org/html/rfc6750#section-2.1)
func TokenFromHeader(r *http.Request) string {
	hdr := r.Header.Get("Authorization")
	if len(hdr) > 7 && hdr[0:7] == "Bearer " {
		return hdr[7:]
	}
	return ""
}


// MustToken retrieves a token from a HTTP request, validating signature and claims.
// It passes control to the next handler in case of success, or returns an error if validation fails.
func (jh *JwtHandler) MustToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := TokenFromHeader(r)
		if token == "" {
			jh.errorFunc(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		/*
		msg, err := jws.ParseString(token)
		if err != nil {
			log.Println("jws.ParseString:", err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		*/
		
		/*
		msg, err := jws.Verify([]byte(token), jwa.HS256, []byte("secret"))
		if err != nil {
			log.Println("jws.Verify:", err)
			//http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			jsonError(w, "error validating token signature", http.StatusUnauthorized)
			return
		}

		log.Printf("jws: %T; %s\n", msg, msg)

		jwtToken, err := jwt.ParseString(token)
		if err != nil {
			log.Println(err)
			//http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			jsonError(w, "error parsing token", http.StatusUnauthorized)
			return
		}
		if err := jwtToken.Verify(); err != nil {
			log.Println(err)
			//http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			jsonError(w, "error validating token claims", http.StatusUnauthorized)
			return
		}
		*/
		
		//jwtToken, err := jwt.ParseString(token, jwt.WithVerify(SignAlgorithm, []byte("secret")))
		jwtToken, err := jwt.ParseString(token, jwt.WithVerify(jh.signAlgorithm, jh.key))
		if err != nil {
			log.Println(err)
			//jsonError(w, err.Error(), http.StatusUnauthorized)
			jh.TokenError(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if err := jwtToken.Verify(jwt.WithAcceptableSkew(10e9), jwt.WithAudience(jh.audience)); err != nil {
			log.Println(err)
			//http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			//jsonError(w, err.Error(), http.StatusUnauthorized)
			jh.TokenError(w, err.Error(), http.StatusUnauthorized)
			return
		}
		
		// check if `jti` is present manually
		if jh.mustJwtID {
			if id, present := jwtToken.Get(jwt.JwtIDKey); !present {
				jh.TokenError(w, errMissingJwtID.Error(), http.StatusUnauthorized)
				return
			} else {
				log.Printf("JwtID: %s\n", id)
			}
		}
			
		log.Printf("jwt: %+v\n", jwtToken)
		
		h.ServeHTTP(w, r)
	})
}

