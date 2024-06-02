package pgconvert

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/maqdev/cachec/cachec"
)

func WrapDBError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.Join(err, cachec.ErrNotFound)
	}
	return err
}

func WrapCacheError(err error) error {
	if errors.Is(err, cachec.ErrNotFound) {
		return errors.Join(err, pgx.ErrNoRows)
	}
	return err
}
