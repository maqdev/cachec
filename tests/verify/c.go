package verify

import (
	"context"
	"time"

	"github.com/maqdev/cachec/cachec/strategies"

	"github.com/maqdev/cachec/cachec"
	"github.com/maqdev/cachec/tests/gen/dal/example"
	"github.com/maqdev/cachec/tests/gen/dal/example/proto"
	exampleDB "github.com/maqdev/cachec/tests/gen/queries/example"
)

const (
	// Entity + Prefix in a single struct!
	AuthorCachePrefix cachec.KeyPrefix = "a"
	AuthorEntityName                   = "author"
)

var (
	AuthorEntity = cachec.CacheEntity{
		KeyPrefix:  AuthorCachePrefix,
		EntityName: AuthorEntityName,
		TTL:        10 * time.Minute,
	}
)

func NewExampleCache(next example.DAL, cacheClient cachec.CacheClient) example.DAL {
	return &exampleCache{
		next:        next,
		cacheClient: cacheClient,
	}
}

type exampleCache struct {
	next        example.DAL
	cacheClient cachec.CacheClient
}

func (e *exampleCache) GetAuthorsByIDs(ctx context.Context, ids []int64) ([]exampleDB.Author, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exampleCache) UpdateAuthor(ctx context.Context, arg exampleDB.UpdateAuthorParams) error {
	//TODO implement me
	panic("implement me")
}

func (e *exampleCache) CreateBook(ctx context.Context, arg exampleDB.CreateBookParams) (exampleDB.Book, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exampleCache) DeleteBook(ctx context.Context, id int64) error {
	//TODO implement me
	panic("implement me")
}

func (e *exampleCache) GetAuthorBooks(ctx context.Context, authorID int64) ([]exampleDB.Book, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exampleCache) GetBook(ctx context.Context, id int64) (exampleDB.Book, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exampleCache) GetBooksByIDs(ctx context.Context, ids []int64) ([]exampleDB.Book, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exampleCache) UpdateBook(ctx context.Context, arg exampleDB.UpdateBookParams) error {
	//TODO implement me
	panic("implement me")
}

func (e *exampleCache) CreateAuthor(ctx context.Context, arg exampleDB.CreateAuthorParams) (exampleDB.Author, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exampleCache) DeleteAuthor(ctx context.Context, id int64) error {
	//TODO implement me
	panic("implement me")
}

func (e *exampleCache) GetAllAuthors(ctx context.Context) ([]exampleDB.Author, error) {
	//TODO implement me
	panic("implement me")
}

func (e *exampleCache) GetAuthor(ctx context.Context, id int64) (exampleDB.Author, error) {
	key := cachec.Key{
		ClusteringKey: &proto.Author__Key{
			ID: id,
		},
	}

	result, err := strategies.GetFromCacheOrNext[exampleDB.Author, proto.Author](ctx, e.cacheClient, "GetAuthor", AuthorEntity, key,
		true, false,
		func() (exampleDB.Author, error) {
			return e.next.GetAuthor(ctx, id)
		},
		proto.AuthorFromPG)

	if err != nil {
		return exampleDB.Author{}, err
	}

	return *result, nil
}

var _ example.DAL = &exampleCache{}
