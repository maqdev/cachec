package verify

import (
	"context"
	"fmt"
	exampleDAL "github.com/maqdev/cachec/tests/gen/dal/example"
	"github.com/maqdev/cachec/tests/gen/dal/example/cache"
	exampleDB "github.com/maqdev/cachec/tests/gen/queries/example"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"testing"
)

func Test_ExampleDAL(t *testing.T) {
	var dbtx exampleDB.DBTX
	db := exampleDB.New(dbtx)

	var dal exampleDAL.DAL
	dal = db

	_, err := dal.GetAuthor(context.Background(), 1)
	require.NoError(t, err)
}

func Test_Encoding(t *testing.T) {
	var a cache.Author
	a.ID = 1

	out, err := proto.Marshal(&a)
	require.NoError(t, err)
	fmt.Println(out)
}
