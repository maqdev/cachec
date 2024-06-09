package strategies

import (
	"context"

	"github.com/maqdev/cachec/cachec"
)

func MGetFromCacheOrNext[PGStruct any, CacheEntityStruct any, CacheEntityIntf CachedEntity[PGStruct, CacheEntityStruct]](
	ctx context.Context, cacheClient cachec.CacheClient, opName string,
	entity cachec.CacheEntity, keys []cachec.Key,
	next func(indices []int) ([]*PGStruct, error),
	convertToCache func(in *PGStruct) CacheEntityIntf) ([]*PGStruct, error) {

	panic("implement me")

	//cacheResult, err := cacheClient.MGet(ctx, entity, keys, func() proto.Message {
	//	var c CacheEntityStruct
	//	return CacheEntityIntf(&c)
	//})
	//
	//// MGet failed in general
	//if err != nil {
	//	return nil, fmt.Errorf("%s/cacheClient.MGet failed: %w", opName, err)
	//}
	//
	//result := make([]*PGStruct, len(keys))
	//var getNext []int
	//
	//for index, v := range cacheResult {
	//	if v.Err != nil {
	//		err = pgconvert.WrapCacheError(v.Err)
	//	}
	//
	//	switch {
	//	// found in cache
	//	case err == nil:
	//		result[index] = v.Message.(CacheEntityIntf).ToPG()
	//
	//	// flagged as not found in cache, left it nil
	//	case errors.Is(err, cachec.ErrNotFound): // do nothing
	//
	//	// other error
	//	case !errors.Is(err, cachec.ErrNotCached):
	//		return nil, fmt.Errorf("%s/cacheClient.MGet failed: %w", opName, err)
	//
	//	// not found in cache, load from the next DAL
	//	default:
	//		if getNext == nil {
	//			getNext = make([]int, 0, len(keys))
	//		}
	//		getNext = append(getNext, index)
	//
	//		result, err := next()
	//		if err != nil {
	//			err = pgconvert.WrapDBError(err)
	//
	//			if errors.Is(err, cachec.ErrNotFound) {
	//				if storeInCacheNotFound {
	//					if !storeInCacheAsynchronously {
	//						// if cacheNotFound is enabled, flag as not found in cache
	//						err = cacheClient.FlagAsNotFound(ctx, entity, key)
	//						if err != nil {
	//							return nil, fmt.Errorf("%s/cacheClient.FlagAsNotFound failed: %w", opName, err)
	//						}
	//					} else {
	//						go func() {
	//							ctxAsync, cancel := context.WithTimeout(context.WithoutCancel(ctx), AsynchronousCacheTimeout)
	//							defer cancel()
	//
	//							err = cacheClient.FlagAsNotFound(ctxAsync, entity, key)
	//							if err != nil {
	//								slog.Error("failed to flag as not found in cache %s: %s", entity.EntityName, err)
	//							}
	//						}()
	//					}
	//				}
	//
	//				return nil, fmt.Errorf("`%s` is not found: %w", entity.EntityName, err)
	//			}
	//		}
	//
	//		newCachedResult := convertToCache(&result)
	//
	//		if !storeInCacheAsynchronously {
	//			// cache asynchronously if strategy allows
	//			err = pgconvert.WrapCacheError(cacheClient.MultiSet(ctx, entity, key, newCachedResult))
	//			if err != nil {
	//				return nil, fmt.Errorf("%s/cacheClient.MultiSet failed: %w", opName, err)
	//			}
	//		} else {
	//			go func() {
	//				ctxAsync, cancel := context.WithTimeout(context.WithoutCancel(ctx), AsynchronousCacheTimeout)
	//				defer cancel()
	//
	//				err = pgconvert.WrapCacheError(cacheClient.MultiSet(ctxAsync, entity, key, newCachedResult))
	//				if err != nil {
	//					slog.Error("failed to cache %s: %s", entity.EntityName, err)
	//				}
	//			}()
	//		}
	//
	//		return &result, nil
	//	}
	//
	//}
	//
	//switch {
	//// found in cache
	//case err == nil:
	//	return cachedResult.ToPG(), nil
	//
	//// flagged as not found in cache
	//case errors.Is(err, cachec.ErrNotFound):
	//	return nil, fmt.Errorf("`%s` is not found: %w", entity.EntityName, err)
	//
	//// other error
	//case !errors.Is(err, cachec.ErrNotCached):
	//	return nil, fmt.Errorf("%s/cacheClient.MultiGet failed: %w", opName, err)
	//
	//// not found in cache, load from the next DAL
	//default:
	//	result, err := next()
	//	if err != nil {
	//		err = pgconvert.WrapDBError(err)
	//
	//		if errors.Is(err, cachec.ErrNotFound) {
	//			if storeInCacheNotFound {
	//				if !storeInCacheAsynchronously {
	//					// if cacheNotFound is enabled, flag as not found in cache
	//					err = cacheClient.FlagAsNotFound(ctx, entity, key)
	//					if err != nil {
	//						return nil, fmt.Errorf("%s/cacheClient.FlagAsNotFound failed: %w", opName, err)
	//					}
	//				} else {
	//					go func() {
	//						ctxAsync, cancel := context.WithTimeout(context.WithoutCancel(ctx), AsynchronousCacheTimeout)
	//						defer cancel()
	//
	//						err = cacheClient.FlagAsNotFound(ctxAsync, entity, key)
	//						if err != nil {
	//							slog.Error("failed to flag as not found in cache %s: %s", entity.EntityName, err)
	//						}
	//					}()
	//				}
	//			}
	//
	//			return nil, fmt.Errorf("`%s` is not found: %w", entity.EntityName, err)
	//		}
	//	}
	//
	//	newCachedResult := convertToCache(&result)
	//
	//	if !storeInCacheAsynchronously {
	//		// cache asynchronously if strategy allows
	//		err = pgconvert.WrapCacheError(cacheClient.MultiSet(ctx, entity, key, newCachedResult))
	//		if err != nil {
	//			return nil, fmt.Errorf("%s/cacheClient.MultiSet failed: %w", opName, err)
	//		}
	//	} else {
	//		go func() {
	//			ctxAsync, cancel := context.WithTimeout(context.WithoutCancel(ctx), AsynchronousCacheTimeout)
	//			defer cancel()
	//
	//			err = pgconvert.WrapCacheError(cacheClient.MultiSet(ctxAsync, entity, key, newCachedResult))
	//			if err != nil {
	//				slog.Error("failed to cache %s: %s", entity.EntityName, err)
	//			}
	//		}()
	//	}
	//
	//	return &result, nil
	//}
}
