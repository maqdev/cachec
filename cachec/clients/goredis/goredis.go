package goredis

import (
	"context"
	"errors"
	"fmt"
	"github.com/maqdev/cachec/cachec"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

type GoRedisCache struct {
	client redis.UniversalClient
}

func NewGoRedisCache(client redis.UniversalClient) *GoRedisCache {
	return &GoRedisCache{
		client: client,
	}
}

func (g *GoRedisCache) Get(ctx context.Context, entity cachec.CacheEntity, key cachec.Key, dest proto.Message) error {
	keyString, err := g.SerializeKey(entity, key)
	if err != nil {
		return err
	}

	cachedBody, err := g.client.Get(ctx, keyString).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return cachec.ErrNotCached
		}
		return err
	}

	if cachedBody == "" {
		return cachec.ErrNotFound
	}

	return proto.Unmarshal([]byte(cachedBody), dest)
}

func (g *GoRedisCache) MGet(ctx context.Context, entity cachec.CacheEntity, keys []cachec.Key, dest []proto.Message, creator func() proto.Message) error {
	//TODO implement me
	panic("implement me")
}

func (g *GoRedisCache) Set(ctx context.Context, entity cachec.CacheEntity, key cachec.Key, src proto.Message) error {
	keyString, err := g.SerializeKey(entity, key)
	if err != nil {
		return err
	}

	cacheBody, err := proto.Marshal(src)
	if err != nil {
		return err
	}

	return g.client.Set(ctx, keyString, cacheBody, entity.TTL).Err()
}

func (g *GoRedisCache) FlagAsNotFound(ctx context.Context, entity cachec.CacheEntity, key cachec.Key) error {
	keyString, err := g.SerializeKey(entity, key)
	if err != nil {
		return err
	}

	return g.client.Set(ctx, keyString, "", entity.TTL).Err()
}

func (g *GoRedisCache) Delete(ctx context.Context, entity cachec.CacheEntity, key cachec.Key) error {
	keyString, err := g.SerializeKey(entity, key)
	if err != nil {
		return err
	}
	return g.client.Del(ctx, keyString).Err()
}

func (g *GoRedisCache) SerializeKey(entity cachec.CacheEntity, key cachec.Key) (string, error) {
	// todo: Redis hashtags, escape { }, partition key, separator after prefix?

	//key.ClusteringKey.

	out, err := proto.Marshal(key.ClusteringKey)
	if err != nil {
		return "", fmt.Errorf("SerializeKey failed: %w", err)
	}
	return string(entity.KeyPrefix) + string(out), nil
}

var _ cachec.CacheClient = (*GoRedisCache)(nil)
