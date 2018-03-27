package psql

import (
	"errors"
	"github.com/jackc/pgx"
	"log"
)

var errInvalidJson = errors.New("invalid json")
var errExists = errors.New("id exists")

// json error:
// error: pgx.PgError{Severity:"ERROR", Code:"22P02", Message:"invalid input syntax for type json", Detail:"Token \"hello\" is invalid.", Hint:"", Position:0, InternalPosition:0, InternalQuery:"", Where:"JSON data, line 1: hello...", SchemaName:"", TableName:"", ColumnName:"", DataTypeName:"", ConstraintName:"", File:"json.c", Line:1244, Routine:"report_invalid_token"}
// ERROR: invalid input syntax for type json (SQLSTATE 22P02)

// 23503: foreign_key_violation; 23505: unique_violation
//pgx.PgError{Severity:"ERROR", Code:"23503", Message:"insert or update on table \"passwd\" violates foreign key constraint \"passwd_uid_fkey\"", Detail:"Key (uid)=(053bffbc-c41e-dad4-853b-ea91fc42ea17) is not present in table \"users\".", Hint:"", Position:0, InternalPosition:0, InternalQuery:"", Where:"", SchemaName:"attid", TableName:"passwd", ColumnName:"", DataTypeName:"", ConstraintName:"passwd_uid_fkey", File:"ri_triggers.c", Line:3324, Routine:"ri_ReportViolation"}
//if pgerr.ConstraintName == "unique_username" || pgerr.ConstraintName == "users_email_lc" {
//	ExistsType = true

// handleError catches some psql errors.
func handleError(err error) error {
	if err == nil {
		return nil
	}

	log.Printf("error: %#v (%T)\n", err, err)

	// pgx/postgres error
	if pgerr, ok := err.(pgx.PgError); ok {
		switch pgerr.Code {
		case "22P02":
			return errInvalidJson
		case "23503":
			return errExists
		case "23505":
			return errExists
		}

		return pgerr
	}

	// don't know, just return
	return err
}
