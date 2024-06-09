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

type MGetRecord struct {
	Message proto.Message
	Err     error
}

type MSetRecord struct {
	Key         Key
	Message     proto.Message
	SetNotFound bool
}

type Cache interface {
	Get(ctx context.Context, entity CacheEntity, key Key, dest proto.Message) error
	Set(ctx context.Context, entity CacheEntity, key Key, src proto.Message) error
	FlagAsNotFound(ctx context.Context, entity CacheEntity, keys ...Key) error
	Delete(ctx context.Context, entity CacheEntity, keys ...Key) error
	MGet(ctx context.Context, entity CacheEntity, keys []Key, creator func() proto.Message) ([]MGetRecord, error)
	MSet(ctx context.Context, entity CacheEntity, set []MSetRecord) error
}

type CacheClient interface {
	Cache
	// todo: pipeline?
}

// ErrNotCached means the cache doesn't have a corresponding entity for the key (wasn't cached yet)
var ErrNotCached = errors.New("not cached")

// ErrNotFound means that the entity is not found neither in cache nor in the database
var ErrNotFound = errors.New("not found")
