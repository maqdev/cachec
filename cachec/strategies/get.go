package strategies

import (
	"context"
	"errors"

	"github.com/maqdev/cachec/cachec"
	"github.com/maqdev/cachec/pgconvert"
)

func GetFromCacheOrNext[PGStruct any, CacheEntityStruct any, CacheEntityIntf CachedEntity[PGStruct, CacheEntityStruct]](
	ctx context.Context, cacheClient cachec.CacheClient, opName string,
	entity cachec.CacheEntity, key cachec.Key,
	next func() (PGStruct, error),
	convertToCache func(in *PGStruct) CacheEntityIntf) (*PGStruct, error) {

	result, err := MGetFromCacheOrNext[PGStruct, CacheEntityStruct, CacheEntityIntf](
		ctx, cacheClient, opName, entity, []cachec.Key{key},
		func(_ []int) ([]*PGStruct, error) {
			res, err := next()
			if err != nil {
				err = pgconvert.WrapDBError(err)
				if errors.Is(err, cachec.ErrNotFound) {
					return []*PGStruct{nil}, nil
				}
				return nil, err
			}
			return []*PGStruct{&res}, nil
		},
		convertToCache,
	)

	if err != nil {
		return nil, err
	}

	return result[0], nil
}
