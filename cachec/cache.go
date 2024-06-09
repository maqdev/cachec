package cachec

import (
	"context"
	"errors"
	"time"

	proto "google.golang.org/protobuf/proto"
)

type CacheEntity struct {
	KeyPrefix     KeyPrefix
	EntityName    string
	TTL           time.Duration
	StoreNotFound bool
	CacheAsync    bool
}

type Key struct {
	PartitionKey  proto.Message
	ClusteringKey proto.Message
}

type KeyPrefix string

type GetResult struct {
	Value proto.Message
	Err   error
}

type SetCommand struct {
	Key         Key
	Value       proto.Message
	SetNotFound bool
	Delete      bool
}

type Cache interface {
	MultiGet(ctx context.Context, entity CacheEntity, keys []Key, creator func() proto.Message) ([]GetResult, error)
	MultiSet(ctx context.Context, entity CacheEntity, set []SetCommand) ([]error, error)
}

type CacheClient interface {
	Cache
	// todo: pipeline?
}

// ErrNotCached means the cache doesn't have a corresponding entity for the key (wasn't cached yet)
var ErrNotCached = errors.New("not cached")

// ErrNotFound means that the entity is not found neither in cache nor in the database
var ErrNotFound = errors.New("not found")
