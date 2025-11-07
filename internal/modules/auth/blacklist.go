package auth

import (
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type Blacklist interface {
	Add(jti string, ttl time.Duration)
	Has(jti string) bool
}

type redisBL struct {
	rc *redis.Client
	ns string // key namespace, ör: "bl:jti:"
}

func NewRedisBlacklist(rc *redis.Client, namespace string) Blacklist {
	if namespace == "" {
		namespace = "bl:jti:"
	}
	return &redisBL{rc: rc, ns: namespace}
}

func (b *redisBL) key(jti string) string { return b.ns + jti }

func (b *redisBL) Add(jti string, ttl time.Duration) {
	ctx := context.Background()
	_ = b.rc.Set(ctx, b.key(jti), "1", ttl).Err()
}

func (b *redisBL) Has(jti string) bool {
	ctx := context.Background()
	res, err := b.rc.Exists(ctx, b.key(jti)).Result()
	return err == nil && res == 1
}
