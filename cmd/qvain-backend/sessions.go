package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/NatLibFi/qvain-api/internal/psql"
	"github.com/NatLibFi/qvain-api/internal/sessions"
	"github.com/NatLibFi/qvain-api/internal/shared"
	"github.com/NatLibFi/qvain-api/pkg/metax"
	"github.com/NatLibFi/qvain-api/pkg/models"

	gooidc "github.com/coreos/go-oidc"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
)

const (
	// projects are in `group_names` field
	FairdataTokenProjectKey = "group_names"

	// silly prefix in projects to get rid of
	FairdataTokenProjectPrefix = "fairdata:IDA01:"
)

// MakeSessionHandlerForExternalService is a callback function that creates a session for a user of an external service.
// It creates an application user account if one doesn't exist.
func MakeSessionHandlerForExternalService(mgr *sessions.Manager, db *psql.DB, logger zerolog.Logger, svc string) func(http.ResponseWriter, *http.Request, string, time.Time) error {
	return func(w http.ResponseWriter, r *http.Request, id string, exp time.Time) error {
		logger.Debug().Str("svc", svc).Str("identity", id).Msg("session callback called")
		uid, isNew, err := db.RegisterIdentity(svc, id)
		if err != nil {
			return err
		}
		sid, _ := mgr.NewLoginWithCookie(w, &uid, nil, sessions.WithExpiration(exp))
		logger.Debug().Str("svc", svc).Str("identity", id).Str("uid", uid.String()).Bool("new", isNew).Msg("new session")

		mgr.List(w)
		session, err := mgr.Get(sid)
		uidTest, _ := session.Uid()
		logger.Debug().Str("sid", sid).Str("uid", uidTest.String()).Msg("get session")
		return nil
	}
}

// MakeSessionHandlerForFairdata is a callback function for the OIDC callback handler to glue token data and our own database to create a user session.
// This particular version handles token fields specific to the Fairdata authentication proxy; see also generic version above.
func MakeSessionHandlerForFairdata(mgr *sessions.Manager, db *psql.DB, onLogin loginHook, logger zerolog.Logger, svc string) func(http.ResponseWriter, *http.Request, *oauth2.Token, *gooidc.IDToken) error {
	return func(w http.ResponseWriter, r *http.Request, oauthToken *oauth2.Token, idToken *gooidc.IDToken) error {
		logger.Debug().Str("svc", svc).Str("identity", idToken.Subject).Msg("session callback called")
		uid, isNew, err := db.RegisterIdentity(svc, idToken.Subject)
		if err != nil {
			return err
		}

		// it is somewhat ok if this stays a nil pointer
		var user *models.User

		// clumsy but the only way to go
		var claims struct {
			Name          string   `json:"name"`
			Email         string   `json:"email"`
			EmailVerified bool     `json:"email_verified"`
			Audience      []string `json:"audience"`
			Projects      []string `json:"group_names"`
			Eppn          string   `json:"eppn"`
			Org           string   `json:"schacHomeOrganization"`
			OrgType       string   `json:"schacHomeOrganizationType"`
		}
		if err := idToken.Claims(&claims); err != nil {
			// let user be nil pointer
			logger.Warn().Err(err).Msg("failed to get token claims")
		} else {
			user = &models.User{
				Uid:          uid,
				Identity:     idToken.Subject,
				Service:      svc,
				Name:         claims.Name,
				Email:        claims.Email,
				Organisation: claims.Org,
			}

			// filter project names returned from the token to include only IDA project numbers
			projects := filterOnAndTrimPrefix(claims.Projects, FairdataTokenProjectPrefix)
			if len(projects) > 0 {
				user.Projects = projects
				logger.Debug().Strs("projects", projects).Msg("ida projects in token")
			}
		}

		sid, err := mgr.NewLoginWithCookie(
			w,
			&uid,
			user,
			sessions.WithExpiration(idToken.Expiry),
		)
		if err != nil {
			return err
		}

		// TODO: remove sid
		logger.Info().Str("svc", svc).Str("identity", idToken.Subject).Str("uid", uid.String()).Bool("new", isNew).Str("sid", sid).Msg("new session")
		if onLogin != nil {
			go onLogin(user)
		}

		return nil
	}
}

type loginHook func(*models.User) error

func makeOnFairdataLogin(metax *metax.MetaxService, db *psql.DB, logger zerolog.Logger) loginHook {
	return func(user *models.User) error {
		return shared.Fetch(metax, db, logger, user.Uid, user.Identity)
	}
}

// filterOnAndTrimPrefix filters a string slice in-place, returning only those items matching the given prefix, then trimming it.
func filterOnAndTrimPrefix(in []string, prefix string) []string {
	out := in[:0]
	for _, project := range in {
		if strings.HasPrefix(project, prefix) {
			out = append(out, strings.TrimPrefix(project, prefix))
		}
	}
	return out
}
