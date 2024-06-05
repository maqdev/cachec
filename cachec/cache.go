package cachec

import (
	"context"
	"errors"
	"time"

	proto "google.golang.org/protobuf/proto"
)

type CacheEntity struct {
	KeyPrefix  KeyPrefix
	EntityName string
	TTL        time.Duration
}

type Key struct {
	PartitionKey  proto.Message
	ClusteringKey proto.Message
}

type KeyPrefix string

type Cache interface {
	Get(ctx context.Context, entity CacheEntity, key Key, dest proto.Message) error
	MGet(ctx context.Context, entity CacheEntity, keys []Key, dest []proto.Message, creator func() proto.Message) error
	Set(ctx context.Context, entity CacheEntity, key Key, src proto.Message) error
	FlagAsNotFound(ctx context.Context, entity CacheEntity, key Key) error
	Delete(ctx context.Context, entity CacheEntity, key Key) error
}

type CacheClient interface {
	Cache
	// todo: pipeline?
}

// ErrNotCached means the cache doesn't have a corresponding entity for the key (wasn't cached yet)
var ErrNotCached = errors.New("not cached")

// ErrNotFound means that the entity is not found neither in cache nor in the database
var ErrNotFound = errors.New("not found")
