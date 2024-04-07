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
	SetNotFoundInDB(ctx context.Context, entity Entity, key Key, ttl time.Time) error
	Delete(ctx context.Context, entity Entity, key Key) error
}

type CacheClient interface {
	Cache
	// todo: pipeline?
}

var ErrNotCached = errors.New("not cached")

var ErrNotFound = errors.Join(errors.New("not found"), ErrNotCached)
