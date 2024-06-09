package strategies

import (
	"time"

	"google.golang.org/protobuf/proto"
)

type CachedEntity[PG any, S any] interface {
	proto.Message
	ToPG() *PG
	*S // constrains type argument to struct that implements this interface
}

var AsynchronousCacheTimeout = time.Second * 10
