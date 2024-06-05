package verify

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"time"

	"github.com/maqdev/cachec/cachec"
	"github.com/maqdev/cachec/pgconvert"
	"github.com/maqdev/cachec/tests/gen/dal/example"
	"github.com/maqdev/cachec/tests/gen/dal/example/cache"
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

type CachedEntity[T any] interface {
	proto.Message
	ToPG() *T
}

func Get[T any, C CachedEntity[T]](ctx context.Context, cacheClient cachec.CacheClient, methodName string,
	entity cachec.CacheEntity, key cachec.Key, next func() (T, error), convert func(in *T) C, dest *T) error {
	var cachedResult C
	err := pgconvert.WrapCacheError(cacheClient.Get(ctx, AuthorEntity, key, cachedResult))

	switch {
	// found in cache
	case err == nil:
		*dest = *cachedResult.ToPG()
		return nil

	// flagged as not found in cache
	case errors.Is(err, cachec.ErrNotFound):
		return fmt.Errorf("`%s` is not found: %w", entity.EntityName, err)

	// other error
	case !errors.Is(err, cachec.ErrNotCached):
		return fmt.Errorf("%s/cacheClient.Get failed: %w", methodName, err)

	// not found in cache, load from the next DAL
	default:
		result, err := next()
		if err != nil {
			err = pgconvert.WrapDBError(err)

			if errors.Is(err, cachec.ErrNotFound) {
				// if cacheNotFound is enabled, flag as not found in cache
				err = cacheClient.FlagAsNotFound(ctx, entity, key)
				if err != nil {
					return fmt.Errorf("%s/cacheClient.FlagAsNotFound failed: %w", methodName, err)
				}

				return fmt.Errorf("`%s` is not found: %w", entity.EntityName, err)
			}
		}

		newCachedResult := convert(&result)

		// cache asynchronously if strategy allows
		err = pgconvert.WrapCacheError(cacheClient.Set(ctx, AuthorEntity, key, newCachedResult))
		if err != nil {
			return fmt.Errorf("%s/cacheClient.Set failed: %w", methodName, err)
		}

		*dest = result
		return nil
	}
}

func (e *exampleCache) GetAuthor(ctx context.Context, id int64) (exampleDB.Author, error) {
	key := cachec.Key{
		ClusteringKey: &cache.Author__Key{
			ID: id,
		},
	}

	var result exampleDB.Author
	err := Get[exampleDB.Author, *cache.Author](ctx, e.cacheClient, "GetAuthor", AuthorEntity, key,
		func() (exampleDB.Author, error) {
			return e.next.GetAuthor(ctx, id)
		},
		cache.AuthorFromPG, &result)

	return result, err
}

/*
func (e *exampleCache) GetAuthor(ctx context.Context, id int64) (exampleDB.Author, error) {
	var cachedResult cache.Author
	key := cachec.Key{
		ClusteringKey: &cache.Author__Key{
			ID: id,
		},
	}
	err := pgconvert.WrapCacheError(e.cacheClient.Get(ctx, AuthorEntity, key, &cachedResult))

	switch {
	// found in cache
	case err == nil:
		return *cachedResult.ToPG(), nil

	// flagged as not found in cache
	case errors.Is(err, cachec.ErrNotFound):
		return exampleDB.Author{}, fmt.Errorf("`Author` is not found: %w", err)

	// other error
	case !errors.Is(err, cachec.ErrNotCached):
		return exampleDB.Author{}, fmt.Errorf("GetAuthor/cacheClient.Get failed: %w", err)

	// not found in cache, load from the next DAL
	default:
		var result exampleDB.Author
		result, err = e.next.GetAuthor(ctx, id)
		if err != nil {
			err = pgconvert.WrapDBError(err)

			if errors.Is(err, cachec.ErrNotFound) {
				// if cacheNotFound is enabled, flag as not found in cache
				err = e.cacheClient.FlagAsNotFound(ctx, AuthorEntity, key)
				if err != nil {
					return exampleDB.Author{}, fmt.Errorf("GetAuthor/cacheClient.FlagAsNotFound failed: %w", err)
				}

				return exampleDB.Author{}, fmt.Errorf("`Author` is not found: %w", err)
			}
		}

		newCachedResult := cache.AuthorFromPG(&result)

		// cache asynchronously if strategy allows
		err = pgconvert.WrapCacheError(e.cacheClient.Set(ctx, AuthorEntity, key, newCachedResult))
		if err != nil {
			return exampleDB.Author{}, fmt.Errorf("GetAuthor/cacheClient.Set failed: %w", err)
		}

		return result, nil
	}
}
*/

var _ example.DAL = &exampleCache{}
