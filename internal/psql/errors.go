package psql

import (
	"log"
	"net"

	"github.com/jackc/pgx"
)

// New returns an error that formats as the given text.
func NewError(text string) error {
	return &DatabaseError{text}
}

// DatabaseError is a trivial implementation of error.
type DatabaseError struct {
	s string
}

// Error satisfies Go's Error interface.
func (e *DatabaseError) Error() string {
	return e.s
}

// Errors exported by the database layer.
var (
	ErrExists         = NewError("exists")
	ErrNotFound       = NewError("not found")
	ErrNotOwner       = NewError("not owner")
	ErrInvalidJson    = NewError("invalid json")
	ErrNotImplemented = NewError("not implemented")
)

// Errors from the underlying database connection.
var (
	ErrTemporary  = NewError("temporary database error")
	ErrTimeout    = NewError("database timeout")
	ErrConnection = NewError("database connection error")
)

// handleError catches some psql errors the application should know about and converts them to one of those defined above.
func handleError(err error) error {
	// heh... shortcut this
	if err == nil {
		return nil
	}

	// TODO: remove this
	log.Printf("[DEBUG] error: %s (%#v; %T)\n", err, err, err)

	// no rows
	if err == pgx.ErrNoRows {
		return ErrNotFound
	}

	// pgx/postgres error
	if pgerr, ok := err.(pgx.PgError); ok {
		switch pgerr.Code {
		//case "22P02":
		//	return ErrInvalidJson
		case "23503":
			return ErrExists
		case "23505":
			return ErrExists
		}

		return pgerr
	}

	// net connection error
	if neterr, ok := err.(*net.OpError); ok {
		if neterr.Temporary() {
			return ErrTemporary
		}
		if neterr.Timeout() {
			return ErrTimeout
		}
		return ErrConnection
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
