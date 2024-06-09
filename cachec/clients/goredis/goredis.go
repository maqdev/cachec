package goredis

import (
	"context"
	"errors"
	"fmt"
	"github.com/maqdev/cachec/cachec"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

type CacheClient struct {
	client redis.UniversalClient
}

func NewGoRedisCache(client redis.UniversalClient) *CacheClient {
	return &CacheClient{
		client: client,
	}
}

func (g *CacheClient) Get(ctx context.Context, entity cachec.CacheEntity, key cachec.Key, dest proto.Message) error {
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

func (g *CacheClient) Set(ctx context.Context, entity cachec.CacheEntity, key cachec.Key, src proto.Message) error {
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

func (g *CacheClient) FlagAsNotFound(ctx context.Context, entity cachec.CacheEntity, keys ...cachec.Key) error {
	// todo: use pipelines & mset

	for _, key := range keys {
		keyString, err := g.SerializeKey(entity, key)
		if err != nil {
			return err
		}

		err = g.client.Set(ctx, keyString, "", entity.TTL).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *CacheClient) Delete(ctx context.Context, entity cachec.CacheEntity, keys ...cachec.Key) error {
	// todo: use pipeline
	for _, key := range keys {
		keyString, err := g.SerializeKey(entity, key)
		if err != nil {
			return err
		}
		err = g.client.Del(ctx, keyString).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *CacheClient) MGet(ctx context.Context, entity cachec.CacheEntity, keys []cachec.Key, creator func() proto.Message) ([]cachec.MGetRecord, error) {
	//TODO implement me
	panic("implement me")
}

func (g *CacheClient) MSet(ctx context.Context, entity cachec.CacheEntity, set []cachec.MSetRecord) error {
	//TODO implement me
	panic("implement me")
}

func (g *CacheClient) SerializeKey(entity cachec.CacheEntity, key cachec.Key) (string, error) {
	// todo: Redis hashtags, escape { }, partition key, separator after prefix?

	//key.ClusteringKey.

	out, err := proto.Marshal(key.ClusteringKey)
	if err != nil {
		return "", fmt.Errorf("SerializeKey failed: %w", err)
	}
	return string(entity.KeyPrefix) + string(out), nil
}

var _ cachec.CacheClient = (*CacheClient)(nil)
