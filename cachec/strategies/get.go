package strategies

import (
	"context"
	"errors"
	"fmt"
	"github.com/maqdev/cachec/cachec"
	"github.com/maqdev/cachec/pgconvert"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"time"
)

type CachedEntity[PG any, S any] interface {
	proto.Message
	ToPG() *PG
	*S // constrains type argument to struct that implements this interface
}

func GetFromCacheOrNext[PGStruct any, CacheEntityStruct any, CacheEntityIntf CachedEntity[PGStruct, CacheEntityStruct]](
	ctx context.Context, cacheClient cachec.CacheClient, opName string,
	entity cachec.CacheEntity, key cachec.Key,
	storeInCacheNotFound bool,
	storeInCacheAsynchronously bool,
	next func() (PGStruct, error),
	converToCache func(in *PGStruct) CacheEntityIntf) (*PGStruct, error) {

	var c CacheEntityStruct
	cachedResult := CacheEntityIntf(&c)
	err := pgconvert.WrapCacheError(cacheClient.Get(ctx, entity, key, cachedResult))

	switch {
	// found in cache
	case err == nil:
		return cachedResult.ToPG(), nil

	// flagged as not found in cache
	case errors.Is(err, cachec.ErrNotFound):
		return nil, fmt.Errorf("`%s` is not found: %w", entity.EntityName, err)

	// other error
	case !errors.Is(err, cachec.ErrNotCached):
		return nil, fmt.Errorf("%s/cacheClient.Get failed: %w", opName, err)

	// not found in cache, load from the next DAL
	default:
		result, err := next()
		if err != nil {
			err = pgconvert.WrapDBError(err)

			if errors.Is(err, cachec.ErrNotFound) {
				if storeInCacheNotFound {
					if !storeInCacheAsynchronously {
						// if cacheNotFound is enabled, flag as not found in cache
						err = cacheClient.FlagAsNotFound(ctx, entity, key)
						if err != nil {
							return nil, fmt.Errorf("%s/cacheClient.FlagAsNotFound failed: %w", opName, err)
						}
					} else {
						go func() {
							ctxAsync, cancel := context.WithTimeout(context.WithoutCancel(ctx), AsynchronousCacheTimeout)
							defer cancel()

							err = cacheClient.FlagAsNotFound(ctxAsync, entity, key)
							if err != nil {
								slog.Error("failed to flag as not found in cache %s: %s", entity.EntityName, err)
							}
						}()
					}
				}

				return nil, fmt.Errorf("`%s` is not found: %w", entity.EntityName, err)
			}
		}

		newCachedResult := converToCache(&result)

		if !storeInCacheAsynchronously {
			// cache asynchronously if strategy allows
			err = pgconvert.WrapCacheError(cacheClient.Set(ctx, entity, key, newCachedResult))
			if err != nil {
				return nil, fmt.Errorf("%s/cacheClient.Set failed: %w", opName, err)
			}
		} else {
			go func() {
				ctxAsync, cancel := context.WithTimeout(context.WithoutCancel(ctx), AsynchronousCacheTimeout)
				defer cancel()

				err = pgconvert.WrapCacheError(cacheClient.Set(ctxAsync, entity, key, newCachedResult))
				if err != nil {
					slog.Error("failed to cache %s: %s", entity.EntityName, err)
				}
			}()
		}

		return &result, nil
	}
}

var AsynchronousCacheTimeout = time.Second * 10
