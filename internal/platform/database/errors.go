package database

import (
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

var (
	ErrNoRows               = pgx.ErrNoRows
	ErrTxDone               = pgx.ErrTxClosed
	UniqueViolationCode     = "23505" //pgx.UniqueViolationCode
	ForeignKeyViolationCode = "23503"
	CheckViolationCode      = "23514"
)

// IsNoRowsError checks if the error is a "no rows" error
func IsNoRowsError(err error) bool {
	return errors.Is(err, ErrNoRows)
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == UniqueViolationCode
	}
	return false
}
