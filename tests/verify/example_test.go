package verify

import (
	"context"
	exampleDAL "github.com/maqdev/cachec/tests/gen/dal/example"
	exampleDB "github.com/maqdev/cachec/tests/gen/queries/example"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ExampleDAL(t *testing.T) {
	var dbtx exampleDB.DBTX
	db := exampleDB.New(dbtx)

	var dal exampleDAL.DAL
	dal = db

	_, err := dal.GetAuthor(context.TODO(), 1)
	require.NoError(t, err)
}
