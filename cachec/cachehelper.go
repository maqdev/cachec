package cachec

import (
	"context"

	"google.golang.org/protobuf/proto"
)

type CacheClientHelper struct {
	CacheClient
}

func (c *CacheClientHelper) Get(ctx context.Context, entity CacheEntity, key Key, dest proto.Message) error {
	panic("implement me")
}

func (c *CacheClientHelper) Set(ctx context.Context, entity CacheEntity, key Key, src proto.Message) error {
	panic("implement me")
}

func (c *CacheClientHelper) FlagAsNotFound(ctx context.Context, entity CacheEntity, keys ...Key) error {
	panic("implement me")
}

func (c *CacheClientHelper) Delete(ctx context.Context, entity CacheEntity, keys ...Key) error {
	panic("implement me")
}
