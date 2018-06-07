package psql

import (
	"errors"
	"github.com/jackc/pgx"
	"log"
)

// Errors exported by the database layer.
var (
	ErrExists      = errors.New("exists")
	ErrNotFound    = errors.New("not found")
	ErrNotOwner    = errors.New("not owner")
	ErrInvalidJson = errors.New("invalid json")
)

// handleError catches some psql errors the application should know about and converts them to one of those defined above.
func handleError(err error) error {
	// heh... shortcut this
	if err == nil {
		return nil
	}

	// TODO: remove this
	log.Printf("[DEBUG] error: %#v (%T)\n", err, err)

	// no rows
	if err == pgx.ErrNoRows {
		return ErrNotFound
	}

	// pgx/postgres error
	if pgerr, ok := err.(pgx.PgError); ok {
		switch pgerr.Code {
		case "22P02":
			return ErrInvalidJson
		case "23503":
			return ErrExists
		case "23505":
			return ErrExists
		}

		return pgerr
	}

	// some other error, just pass on
	return err
}

/*
	Errors good to catch are:

	json error:
		pgx.PgError{Severity:"ERROR", Code:"22P02", Message:"invalid input syntax for type json", Detail:"Token \"hello\" is invalid.", Hint:"", Position:0, InternalPosition:0, InternalQuery:"", Where:"JSON data, line 1: hello...", SchemaName:"", TableName:"", ColumnName:"", DataTypeName:"", ConstraintName:"", File:"json.c", Line:1244, Routine:"report_invalid_token"}
		ERROR: invalid input syntax for type json (SQLSTATE 22P02)

	exists errors (23503: foreign_key_violation; 23505: unique_violation):
		pgx.PgError{Severity:"ERROR", Code:"23503", Message:"insert or update on table \"passwd\" violates foreign key constraint \"passwd_uid_fkey\"", Detail:"Key (uid)=(053bffbc-c41e-dad4-853b-ea91fc42ea17) is not present in table \"users\".", Hint:"", Position:0, InternalPosition:0, InternalQuery:"", Where:"", SchemaName:"attid", TableName:"passwd", ColumnName:"", DataTypeName:"", ConstraintName:"passwd_uid_fkey", File:"ri_triggers.c", Line:3324, Routine:"ri_ReportViolation"}
*/
