package main

import (
	"net/http"

	"github.com/NatLibFi/qvain-api/jwt"
	"github.com/NatLibFi/qvain-api/psql"
	"github.com/rs/zerolog"
	"github.com/wvh/uuid"
)

type Views struct {
	db     *psql.DB
	logger zerolog.Logger
}

func (view *Views) ByOwner() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			user uuid.UUID
			err  error
		)

		token, ok := jwt.FromContext(r.Context())
		if ok {
			user, err = uuid.FromString(token.Subject())
			if err != nil {
				ok = false
			}
		}
		if !ok {
			jsonError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		user = uuid.MustFromString("12345678-9012-3456-7890-123456789012")

		jsondata, err := view.db.ViewByOwner(user)
		if err != nil {
			view.logger.Error().Err(err).Msg("database error")
			jsonError(w, "database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Write(jsondata)
	}
}
