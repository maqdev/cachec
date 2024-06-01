package verify

import (
	"context"
	"errors"
	"fmt"
	"github.com/maqdev/cachec/cachec"
	"github.com/maqdev/cachec/tests/gen/dal/example"
	"github.com/maqdev/cachec/tests/gen/dal/example/cache"
	exampleDB "github.com/maqdev/cachec/tests/gen/queries/example"
	"time"
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
	}
)

func NewExampleCache(next example.DAL, cacheClient cachec.CacheClient) example.DAL {
	return &exampleCache{}
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
	var cachedResult cache.Author
	// how to get key from id?
	key := cachec.Key{}
	err := e.cacheClient.Get(ctx, AuthorEntity, key, &cachedResult)

	switch {
	// found in cache
	case err == nil:
		return *cachedResult.ToPG(), nil

	// flagged as not found in cache
	case errors.Is(err, cachec.ErrNotFound):
		return exampleDB.Author{}, fmt.Errorf("Author is not found: %w", err)

	// other error
	case !errors.Is(err, cachec.ErrNotCached):
		return exampleDB.Author{}, fmt.Errorf("GetAuthor/cacheClient.Get failed: %w", err)

	// not found in cache, load from the next DAL
	default:
		var result exampleDB.Author
		result, err = e.next.GetAuthor(ctx, id)
		if err != nil {
			// todo: convert pg no rows error to ErrNotFound

			if errors.Is(err, cachec.ErrNotFound) {
				// if cacheNotFound is enabled, flag as not found in cache
				// ttl from config
				ttl := time.Now().Add(24 * time.Hour)

				err = e.cacheClient.FlagAsNotFound(ctx, AuthorEntity, key, ttl)
				if err != nil {
					return exampleDB.Author{}, fmt.Errorf("GetAuthor/cacheClient.FlagAsNotFound failed: %w", err)
				}

				return exampleDB.Author{}, fmt.Errorf("Author is not found: %w", err)
			}
		}

		newCachedResult := cache.AuthorFromPG(&result)

		// ttl from config
		ttl := time.Now().Add(24 * time.Hour)

		// cache asynchronously if strategy allows
		err = e.cacheClient.Set(ctx, AuthorEntity, key, newCachedResult, ttl)
		if err != nil {
			return exampleDB.Author{}, fmt.Errorf("GetAuthor/cacheClient.Set failed: %w", err)
		}

		return result, nil
	}
}

var _ example.DAL = &exampleCache{}