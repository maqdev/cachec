package verify

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maqdev/cachec/cachec"
	"github.com/maqdev/cachec/cachec/clients/goredis"
	"github.com/maqdev/cachec/tests/gen/dal/example"
	exampleDB "github.com/maqdev/cachec/tests/gen/queries/example"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ExampleDAL(t *testing.T) {
	db := &mockExampleDAL{}

	redisClient := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: []string{"localhost:6379", "localhost:6379"},
	})

	cacheClient := goredis.NewGoRedisCache(redisClient)

	dal := NewExampleCache(db, cacheClient)

	a, err := dal.GetAuthor(context.Background(), 1)
	require.NoError(t, err)
	fmt.Println("got author", a)
}

type mockExampleDAL struct {
}

func (m mockExampleDAL) CreateAuthor(ctx context.Context, arg exampleDB.CreateAuthorParams) (exampleDB.Author, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockExampleDAL) DeleteAuthor(ctx context.Context, id int64) error {
	//TODO implement me
	panic("implement me")
}

func (m mockExampleDAL) GetAllAuthors(ctx context.Context) ([]exampleDB.Author, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockExampleDAL) GetAuthor(ctx context.Context, id int64) (exampleDB.Author, error) {
	if id == 1 {
		return exampleDB.Author{
			ID:   id,
			Name: "John Smith",
			Bio: pgtype.Text{
				String: "A man was born, lived, died",
				Valid:  true,
			},
		}, nil
	}
	return exampleDB.Author{}, cachec.ErrNotFound
}

func (m mockExampleDAL) GetAuthorsByIDs(ctx context.Context, ids []int64) ([]exampleDB.Author, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockExampleDAL) UpdateAuthor(ctx context.Context, arg exampleDB.UpdateAuthorParams) error {
	//TODO implement me
	panic("implement me")
}

func (m mockExampleDAL) CreateBook(ctx context.Context, arg exampleDB.CreateBookParams) (exampleDB.Book, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockExampleDAL) DeleteBook(ctx context.Context, id int64) error {
	//TODO implement me
	panic("implement me")
}

func (m mockExampleDAL) GetAuthorBooks(ctx context.Context, authorID int64) ([]exampleDB.Book, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockExampleDAL) GetBook(ctx context.Context, id int64) (exampleDB.Book, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockExampleDAL) GetBooksByIDs(ctx context.Context, ids []int64) ([]exampleDB.Book, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockExampleDAL) UpdateBook(ctx context.Context, arg exampleDB.UpdateBookParams) error {
	//TODO implement me
	panic("implement me")
}

var _ example.DAL = mockExampleDAL{}
