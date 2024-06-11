package strategies

import (
	"context"
	"errors"
	"fmt"
	"github.com/maqdev/cachec/pgconvert"
	"google.golang.org/protobuf/proto"
	"log/slog"

	"github.com/maqdev/cachec/cachec"
)

func MGetFromCacheOrNext[PGStruct any, CacheEntityStruct any, CacheEntityIntf CachedEntity[PGStruct, CacheEntityStruct]](
	ctx context.Context, cacheClient cachec.CacheClient, opName string,
	entity cachec.CacheEntity, keys []cachec.Key,
	next func(indices []int) ([]*PGStruct, error),
	convertToCache func(in *PGStruct) CacheEntityIntf) ([]*PGStruct, error) {

	cacheResult, err := cacheClient.MultiGet(ctx, entity, keys, func() proto.Message {
		var c CacheEntityStruct
		return CacheEntityIntf(&c)
	})

	if err != nil {
		return nil, fmt.Errorf("%s/cacheClient.MultiGet failed: %w", opName, err)
	}

	var nextIndices []int
	result := make([]*PGStruct, len(keys))
	for index, v := range cacheResult {
		if v.Err != nil {
			err = pgconvert.WrapCacheError(v.Err)
		} else {
			err = nil
		}

		switch {
		// found in cache
		case err == nil:
			result[index] = v.Value.(CacheEntityIntf).ToPG()

		// flagged as not found in cache, leave it be nil
		case errors.Is(err, cachec.ErrNotFound): // do nothing

		// other error
		case !errors.Is(err, cachec.ErrNotCached):
			return nil, fmt.Errorf("%s/cacheClient.MultiGet failed at %v: %w", opName, keys[index], err)

		// not found in cache, load from the next DAL
		default:
			if nextIndices == nil {
				nextIndices = make([]int, 0, len(keys))
			}
			nextIndices = append(nextIndices, index)
		}
	}

	if len(nextIndices) > 0 {
		var nextResult []*PGStruct
		nextResult, err = next(nextIndices)

		if err != nil {
			err = pgconvert.WrapDBError(err)
			return nil, fmt.Errorf("%s/next failed: %w", opName, err)
		}

		if len(nextResult) != len(nextIndices) {
			return nil, fmt.Errorf("%s/next returned %d results, expected %d", opName, len(nextResult), len(nextIndices))
		}

		MSetCommands := make([]cachec.SetCommand, 0, len(nextIndices))
		for i, index := range nextIndices {
			r := nextResult[i]
			if r != nil {
				MSetCommands = append(MSetCommands, cachec.SetCommand{
					Key:   keys[index],
					Value: convertToCache(r),
				})
				result[index] = r
			} else if entity.StoreNotFound {
				MSetCommands = append(MSetCommands, cachec.SetCommand{
					Key:         keys[index],
					SetNotFound: true,
				})
			}
		}

		if len(MSetCommands) > 0 {
			if entity.CacheAsync {
				go func() {
					ctxAsync, cancel := context.WithTimeout(context.WithoutCancel(ctx), AsynchronousCacheTimeout)
					defer cancel()

					_, err := cacheClient.MultiSet(ctxAsync, entity, MSetCommands)
					if err != nil {
						err = pgconvert.WrapCacheError(err)
						slog.Error("cacheClient.MultiSet failed for %s: %s", entity.EntityName, err)
					}
				}()
			} else {
				_, err = cacheClient.MultiSet(ctx, entity, MSetCommands)
				if err != nil {
					err = pgconvert.WrapCacheError(err)
					return nil, fmt.Errorf("%s/cacheClient.MultiSet failed: %w", opName, err)
				}
			}
		}
	}

	return result, nil
}
