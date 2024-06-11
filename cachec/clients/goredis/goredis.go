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

func (g *CacheClient) MultiGet(ctx context.Context, entity cachec.CacheEntity, keys []cachec.Key, creator func() proto.Message) ([]cachec.GetResult, error) {
	// todo: group keys by hash tags, use pipeline

	redisKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		keyString, err := g.SerializeKey(entity, key)
		if err != nil {
			return nil, fmt.Errorf("SerializeKey failed: %w", err)
		}
		redisKeys = append(redisKeys, keyString)
	}

	redisResult, err := g.client.MGet(ctx, redisKeys...).Result()
	if err != nil {
		return nil, WrapRedisError(err)
	}

	result := make([]cachec.GetResult, len(keys))
	for i, v := range redisResult {
		if v == nil {
			result[i].Err = cachec.ErrNotCached
			continue
		}

		s := v.(string)
		if s == "" {
			result[i].Err = cachec.ErrNotFound
			continue
		}

		val := creator()
		err = proto.Unmarshal([]byte(s), val)
		if err != nil {
			return nil, fmt.Errorf("proto.Unmarshal failed for %v: %w", keys[i], err)
		}
	}
	return result, nil
}

func (g *CacheClient) MultiSet(ctx context.Context, entity cachec.CacheEntity, set []cachec.SetCommand) ([]error, error) {
	// todo: group keys by hash tags, use pipeline
	var delKeys []string

	g.client.MSet()

	g.client.Del()
}

func NewGoRedisCache(client redis.UniversalClient) *CacheClient {
	return &CacheClient{
		client: client,
	}
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

func WrapRedisError(err error) error {
	if errors.Is(err, redis.Nil) {
		return errors.Join(err, cachec.ErrNotCached)
	}
	return err
}
