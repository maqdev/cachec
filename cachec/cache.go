package cachec

import (
	"context"
	"errors"
	"time"

	proto "google.golang.org/protobuf/proto"
)

type Key [][]byte
type Entity string

type Cache interface {
	Get(ctx context.Context, entity Entity, key Key, dest proto.Message) error
	MGet(ctx context.Context, entity Entity, keys []Key, dest []proto.Message, creator func() proto.Message) error
	Set(ctx context.Context, entity Entity, key Key, src proto.Message, ttl time.Time) error
	Delete(ctx context.Context, entity Entity, key Key) error
}

type CacheClient interface {
	Cache
}

var ErrNotFound = errors.New("entity record isn't in cache")
